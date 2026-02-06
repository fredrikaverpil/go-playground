# cloudrun_pubsub

## Quickstart

```bash
echo 'export TF_VAR_project_id="xxx"' > .envrc
direnv allow .
```

```bash
go get cloud.google.com/go/pubsub
docker compose up
```

Go to `http://localhost:8080` and enter a message. See how the Go app's consumer
processes (logs) the message.

## Deploy to GCP using Terraform

Requirements:

- Taskfile
- Terraform
- gcloud

Set the GCP project id:

```bash
echo 'export TF_VAR_project_id="xxx"' > .envrc
direnv allow .
```

Build and push image:

```bash
task build
task push
```

Create infra:

```bash
gcloud auth login
gcloud auth application-default login  # FIX: this should not be needed. There are missing permission in main.tf.

tf init
tf apply
```

Go to Cloud Run, access the URL, post a message. Go to Pub/Sub or Logs Explorer
to see the message being consumed.

To tear it all down:

```bash
tf destroy
task delete
```

## Docs

- https://hub.docker.com/r/google/cloud-sdk
- https://cloud.google.com/go/docs/reference/cloud.google.com/go/pubsub/latest
