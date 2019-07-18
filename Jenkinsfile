pipeline {
  agent {
    kubernetes {
      yaml """
# TODO(lwsanty): labels selectors TBD
#metadata:
#  labels:
#    some-label: bblfsh-performance
#    class: BabelfishPerformanceTest
spec:
  containers:
  - name: bblfsh-performance
    image: bblfsh/performance:latest
    imagePullPolicy: Always
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
    env:
    - name: GOPATH
      value: "/go"
    - name: DRIVER_NAME
      value: "go-driver"
    - name: DRIVER_LANGUAGE
      value: "go"
    - name: DRIVER_LANGUAGE_EXTENSION
      value: ".go"
    - name: DRIVER_REPO
      value: "https://github.com/bblfsh/go-driver.git"
    - name: DRIVER_SRC_TARGET
      value: "/root/driver"
    - name: DRIVER_SRC_FIXTURES
      value: "/root/driver/fixtures"
    - name: BENCHMARK_FILE
      value: "/root/bench.log"
    - name: LOG_LEVEL
      value: "debug"
    - name: PROM_ADDRESS
      value: "http://prom-pushgateway-prometheus-pushgateway.monitoring.svc.cluster.local:9091"
    - name: PROM_JOB
      value: "bblfsh_perfomance"
    command:
    - cat
    tty: true
"""
      }
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
          checkout scm
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
    stage('Get git commit hash') {
       steps {
          script {
            GIT_COMMIT_HASH = sh(script: "git log -n 1 --pretty=format:'%H'", returnStdout: true)
          }
       }
    }
    stage('Store transformations benchmark to prometheus') {
      when { branch 'jenkins-integration' }
      steps {
        sh 'cd ${DRIVER_SRC_TARGET}'
        sh '/bin/bblfsh-performance parse-and-store --language="${DRIVER_LANGUAGE}" --commit="${GIT_COMMIT_HASH}" --storage="prom" "${BENCHMARK_FILE}"'
      }
    }
    stage('Run end-to-end benchmark') {
      when { branch 'jenkins-integration' }
      steps {
        sh '/bin/bblfsh-performance end-to-end --language="${DRIVER_LANGUAGE}" --commit="${GIT_COMMIT_HASH}" --extension="${DRIVER_LANGUAGE_EXTENSION}" --storage="prom" "${DRIVER_SRC_FIXTURES}"'
      }
    }
  }
}
