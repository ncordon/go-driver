pipeline {
  agent {
    kubernetes {
      yaml """
metadata:
  labels:
    some-label: bblfsh-performance
    class: BabelfishPerformanceTest
spec:
  containers:
  - name: bblfsh-performance
    image: bblfsh/performance:latest
    securityContext:
      privileged: true
      runAsUser: 1000
      fsGroup: 1000
    volumeMounts:
    - mountPath: /var/run/docker.sock
      name: docker-socket
    volumes:
    - name: docker-socket
      hostPath:
        path: /var/run/docker.sock
    command:
    - cat
    tty: true
"""
      }
    }
  }
  environment {
    GOPATH = "/go"
    // this name would be specific for each driver,
    // because each driver repo would have it's own Jenkinsfile
    DRIVER_NAME = "go-driver"
    DRIVER_LANGUAGE = "go"
    DRIVER_LANGUAGE_EXTENSION = ".go"
    DRIVER_REPO = "https://github.com/bblfsh/${env.DRIVER_NAME}.git"
    DRIVER_SRC_TARGET = "/root/driver"
    DRIVER_SRC_FIXTURES = "/root/driver/fixtures"
    BENCHMARK_FILE = "/root/bench.log"
    // this section represents envs that required by bblfsh-performance
    LOG_LEVEL = "debug"
    // address of prometheus pushgateway, prometheus pushgateway must be accessible from jenkins
    PROM_ADDRESS="http://prom-pushgateway-prometheus-pushgateway.monitoring.svc.cluster.local:9091"
    // existing job in prometheus
    PROM_JOB=bblfsh_perfomance
  }
  // this is polling for every 2 minutes
  // however it's better to use trigger curl http://yourserver/jenkins/git/notifyCommit?url=<URL of the Git repository>
  // https://kohsuke.org/2011/12/01/polling-must-die-triggering-jenkins-builds-from-a-git-hook/
  // the problem is that it requires Jenkins to be accessible from the hook side
  triggers { pollSCM('H/2 * * * *') }
  stages {
    stage('Pull') {
      when { branch 'jenkins-integration' }
      steps {
        dir('${DRIVER_SRC_TARGET}') {
          git url: '${DRIVER_REPO}'
        }
      }
    }
    stage('Run transformations benchmark') {
      when { branch 'jenkins-integration' }
      steps {
        sh 'cd ${DRIVER_SRC_TARGET}'
        sh 'go test -run=NONE -bench=. ./driver/... | tee ${BENCHMARK_FILE}'
      }
    }
    stage('Store transformations benchmark to prometheus') {
      when { branch 'jenkins-integration' }
      steps {
        sh 'cd ${DRIVER_SRC_TARGET}'
        GIT_COMMIT_HASH = sh (script: "git log -n 1 --pretty=format:'%H'", returnStdout: true)
        sh '/bin/bblfsh-performance parse-and-store --language="${DRIVER_LANGUAGE}" --commit="${GIT_COMMIT_HASH}" --storage="prom" "${BENCHMARK_FILE}"'
      }
    }
    stage('Run end-to-end benchmark') {
      when { branch 'jenkins-integration' }
      steps {
        GIT_COMMIT_HASH = sh (script: "git log -n 1 --pretty=format:'%H'", returnStdout: true)
        sh '/bin/bblfsh-performance end-to-end --language="${DRIVER_LANGUAGE}" --commit="${GIT_COMMIT_HASH}" --extension="${DRIVER_LANGUAGE_EXTENSION}" --storage="prom" "${DRIVER_SRC_FIXTURES}"'
      }
    }
  }
}
