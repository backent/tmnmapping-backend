pipeline {
    agent any

g    stages {

        // ── 1. Checkout ───────────────────────────────────────────────────────
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        // ── 2. Guard – skip builds triggered by the version-bump commit ───────
        stage('Check Skip CI') {
            steps {
                script {
                    def commitMsg = sh(script: 'git log -1 --pretty=%B', returnStdout: true).trim()
                    if (commitMsg.contains('[skip ci]')) {
                        currentBuild.result = 'NOT_BUILT'
                        error('Version-bump commit detected – skipping pipeline.')
                    }
                }
            }
        }

        // ── 3. Read Version & Config ──────────────────────────────────────────
        stage('Read Version & Config') {
            steps {
                script {
                    env.APP_VERSION = readFile('VERSION').trim()
                    // Credential ID: "tmn-app-backend-image-name" (Secret text) – Docker image repo (e.g. account/image-name)
                    withCredentials([string(credentialsId: 'tmn-app-backend-image-name', variable: 'IMG')]) {
                        env.IMAGE_NAME = IMG
                    }
                    echo "Building version: ${env.APP_VERSION}"
                }
            }
        }

        // ── 4. Build & Test (parallel) ────────────────────────────────────────
        stage('Build & Test') {
            parallel {

                stage('Build Image') {
                    steps {
                        // Dockerfile is at repo root; build context is .
                        sh "docker build -t ${env.IMAGE_NAME}:${env.APP_VERSION} . --platform=linux/amd64"
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

        // ── 5. Push Image ─────────────────────────────────────────────────────
        stage('Push Image') {
            steps {
                // Credential ID: "docker-backent-cred"
                // Type         : Username with Password
                // Username     : Docker Hub username (backent)
                // Password     : Docker Hub password / access token
                withCredentials([usernamePassword(
                    credentialsId: 'docker-backent-cred',
                    usernameVariable: 'DOCKER_USER',
                    passwordVariable: 'DOCKER_PASS'
                )]) {
                    sh 'echo "$DOCKER_PASS" | docker login -u "$DOCKER_USER" --password-stdin'
                    sh "docker push ${env.IMAGE_NAME}:${env.APP_VERSION}"
                }
            }
        }

        // ── 6. Deploy to Server ───────────────────────────────────────────────
        stage('Deploy to Server') {
            steps {
                // Credential IDs (all "Secret text"):
                //   tmn-app-server-host  – SSH server address (e.g. 247.247.247.247)
                //   tmn-app-server-user  – SSH username       (e.g. ubuntu)
                //   tmn-app-server-port  – SSH port           (e.g. 22)
                // Credential ID: "tmn-app-ssh-key"
                //   Type        : SSH Username with private key
                //   Private key : contents of tmn-app-key.pem
                withCredentials([
                    sshUserPrivateKey(credentialsId: 'tmn-app-ssh-key', keyFileVariable: 'SSH_KEY_FILE'),
                    string(credentialsId: 'tmn-app-server-host', variable: 'SERVER_HOST'),
                    string(credentialsId: 'tmn-app-server-user', variable: 'SERVER_USER'),
                    string(credentialsId: 'tmn-app-server-port', variable: 'SERVER_PORT')
                ]) {
                    sh """
                        ssh -i \$SSH_KEY_FILE -o StrictHostKeyChecking=no -p \$SERVER_PORT \$SERVER_USER@\$SERVER_HOST '
                            sudo docker pull ${env.IMAGE_NAME}:${env.APP_VERSION} &&
                            sudo docker rm -f backend || true &&
                            sudo docker run -dp 127.0.0.1:8080:8088 \\
                                --env-file .env \\
                                --network global-network \\
                                --name backend \\
                                --restart unless-stopped \\
                                ${env.IMAGE_NAME}:${env.APP_VERSION}
                        '
                    """
                }
            }
        }

        // ── 7. Tag Release & Bump Minor Version ───────────────────────────────
        stage('Tag & Bump Version') {
            steps {
                // Credential ID: "github-cred"
                // Type         : Username with Password
                // Username     : GitHub username
                // Password     : GitHub personal access token (with repo + write:packages scope)
                withCredentials([usernamePassword(
                    credentialsId: 'tmn-mapping-github-cred',
                    usernameVariable: 'GIT_USER',
                    passwordVariable: 'GIT_TOKEN'
                )]) {
                    script {
                        def currentVersion = env.APP_VERSION

                        // Calculate next minor version  (major.minor+1.0)
                        def parts   = currentVersion.tokenize('.')
                        def major   = parts[0]
                        def minor   = parts[1].toInteger() + 1
                        def nextVersion = "${major}.${minor}.0"

                        // Detect current branch (works for both regular and multibranch jobs)
                        def branch = env.BRANCH_NAME ?: sh(
                            script: 'git rev-parse --abbrev-ref HEAD',
                            returnStdout: true
                        ).trim()

                        // Build authenticated remote URL (handles both HTTPS and SSH origins)
                        sh """
                            ORIGIN_URL=\$(git config --get remote.origin.url)

                            # Normalise SSH → HTTPS  (git@github.com:owner/repo.git  →  https://github.com/owner/repo.git)
                            if echo "\$ORIGIN_URL" | grep -qE '^git@'; then
                                ORIGIN_URL=\$(echo "\$ORIGIN_URL" | sed -E 's|git@([^:]+):|https://\\1/|')
                            fi

                            # Strip any existing credentials embedded in the URL
                            ORIGIN_URL=\$(echo "\$ORIGIN_URL" | sed -E 's|https://[^@]+@|https://|')

                            AUTH_URL=\$(echo "\$ORIGIN_URL" | sed "s|https://|https://\${GIT_USER}:\${GIT_TOKEN}@|")

                            git config user.email "jenkins@ci.local"
                            git config user.name  "Jenkins CI"

                            # Create annotated git tag for the version we just deployed
                            git tag -a "v${currentVersion}" -m "Release v${currentVersion}"

                            # Write the next version and commit with [skip ci] to prevent a loop
                            echo "${nextVersion}" > VERSION
                            git add VERSION
                            git commit -m "ci: bump version to ${nextVersion} [skip ci]"

                            # Push the version-bump commit and the new tag
                            git push "\$AUTH_URL" HEAD:${branch}
                            git push "\$AUTH_URL" "v${currentVersion}"
                        """

                        echo "Tagged v${currentVersion} and bumped VERSION to ${nextVersion}"
                    }
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
            echo "Deployed ${env.IMAGE_NAME}:${env.APP_VERSION} successfully."
        }
        failure {
            echo "Pipeline failed. Image ${env.IMAGE_NAME}:${env.APP_VERSION} was NOT deployed."
        }
    }
}
