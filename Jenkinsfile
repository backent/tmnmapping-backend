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
                        sh "docker build -t ${IMAGE_NAME}:${params.IMAGE_TAG} . --platform=linux/amd64"
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

        // ── 3. Integration Test ───────────────────────────────────────────────
        stage('Integration Test') {
            steps {
                // Start the PostGIS container, create the tmn_test database,
                // run all integration tests, then always tear down the container.
                sh '''
                    docker compose -f docker-compose.postgres.yml up -d

                    echo "Waiting for PostgreSQL to be ready..."
                    for i in $(seq 1 30); do
                        docker compose -f docker-compose.postgres.yml exec -T postgres \
                            pg_isready -U postgres && break
                        sleep 2
                    done

                    docker compose -f docker-compose.postgres.yml exec -T postgres \
                        psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'tmn_test'" | grep -q 1 || \
                        docker compose -f docker-compose.postgres.yml exec -T postgres \
                            psql -U postgres -c "CREATE DATABASE tmn_test;"

                    docker run --rm \
                        --network host \
                        -v "$(pwd):/app" \
                        -w /app \
                        -e POSTGRES_HOST=localhost \
                        -e POSTGRES_PORT=5432 \
                        -e POSTGRES_USER=postgres \
                        -e POSTGRES_PASSWORD=adminlocal \
                        -e POSTGRES_DATABASE=tmn_test \
                        -e POSTGRES_SSLMODE=disable \
                        golang:1.23-alpine \
                        sh -c "go mod download && go test -tags integration -timeout 180s ./integration/..."
                '''
            }
            post {
                always {
                    sh 'docker compose -f docker-compose.postgres.yml down || true'
                }
            }
        }

        // ── 4. Push Image ─────────────────────────────────────────────────────
        stage('Push Image') {
            steps {
                // Credential ID: "dockerhub-credentials"
                // Type         : Username with Password
                // Username     : Docker Hub username (backent)
                // Password     : Docker Hub password / access token
                withCredentials([usernamePassword(
                    credentialsId: 'docker-backent-cred',
                    usernameVariable: 'DOCKER_USER',
                    passwordVariable: 'DOCKER_PASS'
                )]) {
                    sh 'echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin'
                    sh "docker push ${IMAGE_NAME}:${params.IMAGE_TAG}"
                }
            }
        }

        // ── 5. Deploy to Server ───────────────────────────────────────────────
        stage('Deploy to Server') {
            steps {
                // Credential ID: "tmn-app-ssh-key"
                // Type         : SSH Username with private key
                // Username     : ubuntu
                // Private key  : contents of tmn-app-key.pem
                withCredentials([sshUserPrivateKey(
                    credentialsId: 'tmn-app-ssh-key',
                    keyFileVariable: 'SSH_KEY_FILE'
                )]) {
                    sh """
                        ssh -i \$SSH_KEY_FILE -o StrictHostKeyChecking=no ${SERVER_USER}@${SERVER_HOST} '
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
