pipeline {
    agent any

    parameters {
        string(
            name: 'IMAGE_TAG',
            defaultValue: '1.5',
            description: 'Docker image tag / version to build and deploy (e.g. 1.5, 1.6)'
        )
    }

    environment {
        IMAGE_NAME  = 'backent/tmn-mapping-backend'
        SERVER_HOST = '108.136.218.247'
        SERVER_USER = 'ubuntu'
    }

    stages {

        // ── 1. Checkout ───────────────────────────────────────────────────────
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        // ── 2. Build & Test (parallel) ────────────────────────────────────────
        stage('Build & Test') {
            parallel {

                stage('Build Image') {
                    steps {
                        // Dockerfile is at repo root; build context is .
                        sh "docker build -t ${IMAGE_NAME}:${params.IMAGE_TAG} ."
                    }
                }

                stage('Unit Test') {
                    steps {
                        sh '''
                            docker run --rm \
                                -v "$(pwd):/app" \
                                -w /app \
                                golang:1.23-alpine \
                                sh -c "ls -la && go mod download && go test ./services/... -v"
                        '''
                    }
                }

            }
        }

        // ── 3. Push Image ─────────────────────────────────────────────────────
        stage('Push Image') {
            steps {
                // Credential ID: "dockerhub-credentials"
                // Type         : Username with Password
                // Username     : Docker Hub username (backent)
                // Password     : Docker Hub password / access token
                withCredentials([usernamePassword(
                    credentialsId: 'dockerhub-credentials',
                    usernameVariable: 'DOCKER_USER',
                    passwordVariable: 'DOCKER_PASS'
                )]) {
                    sh 'echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin'
                    sh "docker push ${IMAGE_NAME}:${params.IMAGE_TAG}"
                }
            }
        }

        // ── 4. Deploy to Server ───────────────────────────────────────────────
        stage('Deploy to Server') {
            steps {
                // Credential ID: "tmn-app-ssh-key"
                // Type         : SSH Username with private key
                // Username     : ubuntu
                // Private key  : contents of tmn-app-key.pem
                sshagent(credentials: ['tmn-app-ssh-key']) {
                    sh """
                        ssh -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_HOST} '
                            sudo docker pull ${IMAGE_NAME}:${params.IMAGE_TAG} &&
                            sudo docker rm -f backend || true &&
                            sudo docker run -dp 127.0.0.1:8080:8088 \\
                                --env-file .env \\
                                --network global-network \\
                                --name backend \\
                                --restart unless-stopped \\
                                ${IMAGE_NAME}:${params.IMAGE_TAG}
                        '
                    """
                }
            }
        }

    }

    // ── Post-pipeline ─────────────────────────────────────────────────────────
    post {
        always {
            // Logout regardless of success or failure
            sh 'docker logout || true'
        }
        success {
            echo "Deployed ${IMAGE_NAME}:${params.IMAGE_TAG} successfully."
        }
        failure {
            echo "Pipeline failed. Image ${IMAGE_NAME}:${params.IMAGE_TAG} was NOT deployed."
        }
    }
}
