#!/bin/bash

echo -e "
service:
  port: ${LISTEN_PORT:-80}

githubConfig:
  accessToken: "${ACCESS_TOKEN}"
  locations:
    - owner: "${REPO_OWNER}"
      name: "${REPO_NAME}"
      path: "${REPO_PATH}"
"
