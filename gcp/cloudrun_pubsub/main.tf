terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 3.5"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = "us-central1"
  zone    = "us-central1-a"
}

resource "google_service_account" "my_service_account" {
  account_id   = "my-app-account"
  display_name = "Service account for My App"
}

resource "google_project_service" "cloud_run_api" {
  service            = "run.googleapis.com"
  disable_on_destroy = true
}

resource "google_project_service" "pubsub_api" {
  service            = "pubsub.googleapis.com"
  disable_on_destroy = true
}

resource "google_cloud_run_service" "default" {
  name     = "my-app-service"
  location = "us-central1"

  depends_on = [google_project_service.cloud_run_api, google_project_service.pubsub_api]

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/app"

        env {
          name  = "TF_VAR_project_id"
          value = var.project_id
        }
      }
      service_account_name = google_service_account.my_service_account.email
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

resource "google_cloud_run_service_iam_member" "invoker" {
  location = google_cloud_run_service.default.location
  project  = google_cloud_run_service.default.project
  service  = google_cloud_run_service.default.name

  role   = "roles/run.invoker"
  member = "allUsers"
}

resource "google_project_iam_member" "pubsub_editor" {
  project = var.project_id
  role    = "roles/pubsub.editor"
  member  = "serviceAccount:${google_service_account.my_service_account.email}"
}

