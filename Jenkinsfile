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
    command:
    - cat
    tty: true
"""
    }
  }
  // TODO use WORKSPACE /home/jenkins/workspace/sh_go-driver_jenkins-integration
  environment {
    GOPATH = "/go"
    PATH="${env.GOPATH}/bin:$PATH"
    DRIVER_NAME = "go-driver"
    DRIVER_LANGUAGE = "go"
    DRIVER_LANGUAGE_EXTENSION = ".go"
    DRIVER_REPO = "https://github.com/bblfsh/go-driver.git"
    DRIVER_SRC_TARGET = "driver"
    DRIVER_SRC_FIXTURES = "driver/fixtures"
    BENCHMARK_FILE = "bench.log"
    LOG_LEVEL = "debug"
    PROM_ADDRESS = "http://prom-pushgateway-prometheus-pushgateway.monitoring.svc.cluster.local:9091"
    PROM_JOB = "bblfsh_perfomance"
  }
  // this is polling for every 2 minutes
  // however it's better to use trigger curl http://yourserver/jenkins/git/notifyCommit?url=<URL of the Git repository>
  // https://kohsuke.org/2011/12/01/polling-must-die-triggering-jenkins-builds-from-a-git-hook/
  // the problem is that it requires Jenkins to be accessible from the hook side
  triggers { pollSCM('H/2 * * * *') }
  stages {
    // stage('Pull') {
    //   when { branch 'jenkins-integration' }
    //   steps {
    //     dir("${env.DRIVER_SRC_TARGET}") {
    //       checkout scm
    //     }
    //   }
    // }
    stage('Run transformations benchmark') {
      when { branch 'jenkins-integration' }
      steps {
        sh "printenv"
        sh "ls ${env.GOPATH}/bin/"
        sh "go version"
        sh "go test -run=NONE -bench=. ./driver/... | tee ${env.BENCHMARK_FILE}"
      }
    }
    stage('Store transformations benchmark to prometheus') {
      when { branch 'jenkins-integration' }
      steps {
        sh "/root/bblfsh-performance parse-and-store --language=${env.DRIVER_LANGUAGE} --commit=${env.GIT_COMMIT} --storage=prom ${env.BENCHMARK_FILE}"
      }
    }
    stage('Run end-to-end benchmark') {
      when { branch 'jenkins-integration' }
      steps {
        sh "/root/bblfsh-performance end-to-end --language=${env.DRIVER_LANGUAGE} --commit=${env.GIT_COMMIT} --extension=${env.DRIVER_LANGUAGE_EXTENSION} --storage=prom ${env.DRIVER_SRC_FIXTURES}"
      }
    }
  }
}
