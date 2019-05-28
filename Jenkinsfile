pipeline {
    agent { 
        docker { 
            image 'sysdiglabs/operator-builder' 
            args '-u root -v /var/run/docker.sock:/var/run/docker.sock'
        }
    }

    parameters { 
        string(name: 'VERSION', defaultValue: '', description: 'Version to create') 
    }

    stages {
        stage('New-Upstream') {
            steps {
                sh 'helm init --client-only'
                sshagent(credentials: ['github-ssh']) {
                    sh "make new-upstream VERSION=${params.VERSION}"
                }
            }
        }
    }
}

