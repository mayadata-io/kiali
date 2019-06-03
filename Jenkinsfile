// vi:set ts=2 shiftwidth=2 et
def ORG = "mayadataio"
def REPO = "kiali"
def DOCKER_HUB_REPO = "https://index.docker.io/v1/"
def BRANCH_NAME = BRANCH_NAME.toLowerCase()
env.user="atulabhi"
env.pass="ka879707"

  pipeline {
    agent {
      label {
        label ""
          customWorkspace "/var/lib/jenkins/workspace/${REPO}-${BRANCH_NAME}"
      }
    }
    options {
      timestamps()
        timeout(time: 15, unit: 'MINUTES')
    }
    stages {
      stage('Version Update') {
        steps {
           echo "Workspace dir is ${pwd()}"
           script {
             if (env.BRANCH_NAME == 'mayadata-io'){
                  BN = sh(
                    returnStdout: true,
                    script: "./version_override ${REPO}"
                  ).trim()
                echo "${BN}"
                echo "This image will be tagged with ${BN}"
              }else {
                echo "This is not a master branch, tagging skipped!!" 
              } 
           }
        }
      }
      stage('Build for maya-io-agent') {
        steps {
          echo "Workspace dir is ${pwd()}"
            script {
              GIT_SHA = sh(returnStdout: true, script: "git log -n 1 --pretty=format:'%h'").trim()
                echo "Checked out branch: ${env.BRANCH_NAME}"
                sh(script: "make build", returnStdout: true)
            }
        }
      }
      stage('Prepare build artifacts directory') {
        steps {
          script {
            if (env.BRANCH_NAME == 'staging' || env.BRANCH_NAME == 'master' || env.BRANCH_NAME.startsWith('alpha-r')) {
                sh(script: "docker login --username=mayabot --password=M4Y4@openebs", returnStdout: true)
                sh(returnStdout: true, script: "mkdir -p ${HOME}/${ORG}/${REPO}/${BRANCH_NAME}/")
                sh(returnStdout: true, script: "rm -f ${HOME}/${ORG}/${REPO}/${BRANCH_NAME}/*")
            }
          }
        }
      }
      stage('Publish image and create a build artifact') {
        steps {
          script {
            if (env.BRANCH_NAME == 'mayadata-io') {
                sh(returnStdout: true, script: "touch ${HOME}/${ORG}/${REPO}/${BRANCH_NAME}/${BRANCH_NAME}-${GIT_SHA}")
                sh(returnStdout: true, script: "docker push ${ORG}/${REPO}:${env.BRANCH_NAME}-${GIT_SHA}")
            }
          }
        }
      }
      stage ('Adding git-tag for master') {
        steps {
          script {
             if (env.BRANCH_NAME == 'master') {
               sh """
                git tag -fa "${BN}" -m "Release of ${BN}"
                """
                sh "git tag -l"
                sh """
                  git push https://${env.user}:${env.pass}@github.com/mayadata-io/maya-io-agent.git --tag
                   """
             }
          }
        }
      }
    }

    post {
      always {
        echo 'This will always run'
          deleteDir()
      }
      success {
        echo 'This will run only if successful'
          slackSend channel: '#jenkins-builds',
            color: 'good',
            message: "The pipeline ${currentBuild.fullDisplayName} completed successfully :dance: :thumbsup: "
                 
      }
      failure {
        echo 'This will run only if failed'
          slackSend channel: '#jenkins-builds',
            color: 'RED',
            message: "The pipeline ${currentBuild.fullDisplayName} failed. :scream_cat: :japanese_goblin: "

      }
    }
  }
