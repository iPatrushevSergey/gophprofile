{{/*
Expand the name of the chart.
*/}}
{{- define "gophprofile.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "gophprofile.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "gophprofile.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "gophprofile.labels" -}}
helm.sh/chart: {{ include "gophprofile.chart" . }}
{{ include "gophprofile.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "gophprofile.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gophprofile.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "gophprofile.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "gophprofile.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Component-specific labels (server, processor, postgres, etc.)
*/}}
{{- define "gophprofile.componentLabels" -}}
{{ include "gophprofile.selectorLabels" . }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
PostgreSQL DSN for server, processor and migrate job.
*/}}
{{- define "gophprofile.databaseURI" -}}
{{- $pg := .Values.postgres -}}
{{- printf "postgres://%s:%s@%s-postgres:5432/%s?sslmode=disable" $pg.user $pg.password (include "gophprofile.fullname" .) $pg.database -}}
{{- end }}

{{/*
RabbitMQ AMQP URL for server and processor.
*/}}
{{- define "gophprofile.brokerURL" -}}
{{- $rmq := .Values.rabbitmq -}}
{{- printf "amqp://%s:%s@%s-rabbitmq:5672/" $rmq.user $rmq.password (include "gophprofile.fullname" .) -}}
{{- end }}

{{/*
Pod-level security context for app workloads.
*/}}
{{- define "gophprofile.podSecurityContext" -}}
runAsNonRoot: true
runAsUser: {{ .runAsUser }}
runAsGroup: {{ .runAsGroup }}
fsGroup: {{ .fsGroup }}
seccompProfile:
  type: RuntimeDefault
{{- end }}

{{/*
Container-level security context.
*/}}
{{- define "gophprofile.containerSecurityContext" -}}
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop:
    - ALL
{{- end }}

{{/*
GOPHPROFILE_* env entries for server/processor containers.
Component-only settings use server.env / processor.env in values.
*/}}
{{- define "gophprofile.containerEnv" -}}
{{- $appSecret := printf "%s-app" (include "gophprofile.fullname" .) -}}
{{- if .Values.externalSecrets.enabled }}
- name: GOPHPROFILE_DATABASE_URI
  valueFrom:
    secretKeyRef:
      name: {{ $appSecret }}
      key: GOPHPROFILE_DATABASE_URI
- name: GOPHPROFILE_BROKER_URL
  valueFrom:
    secretKeyRef:
      name: {{ $appSecret }}
      key: GOPHPROFILE_BROKER_URL
- name: GOPHPROFILE_MINIO_ENDPOINT
  value: {{ printf "%s-minio:9000" (include "gophprofile.fullname" .) | quote }}
- name: GOPHPROFILE_MINIO_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $appSecret }}
      key: GOPHPROFILE_MINIO_ACCESS_KEY
- name: GOPHPROFILE_MINIO_SECRET_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $appSecret }}
      key: GOPHPROFILE_MINIO_SECRET_KEY
{{- else }}
- name: GOPHPROFILE_DATABASE_URI
  value: {{ include "gophprofile.databaseURI" . | quote }}
- name: GOPHPROFILE_BROKER_URL
  value: {{ include "gophprofile.brokerURL" . | quote }}
- name: GOPHPROFILE_MINIO_ENDPOINT
  value: {{ printf "%s-minio:9000" (include "gophprofile.fullname" .) | quote }}
- name: GOPHPROFILE_MINIO_ACCESS_KEY
  value: {{ .Values.minio.rootUser | quote }}
- name: GOPHPROFILE_MINIO_SECRET_KEY
  value: {{ .Values.minio.rootPassword | quote }}
{{- end }}
- name: GOPHPROFILE_TELEMETRY_OTLP_ENDPOINT
  value: {{ printf "%s-otel-collector:4317" (include "gophprofile.fullname" .) | quote }}
- name: GOPHPROFILE_TELEMETRY_OTLP_INSECURE
  value: "true"
- name: USER
  value: gophprofile
{{- end }}

{{/*
PGPASSWORD env for wait-migrate init container (psql auth).
*/}}
{{- define "gophprofile.postgresPasswordEnv" -}}
{{- if .Values.externalSecrets.enabled }}
- name: PGPASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "gophprofile.fullname" . }}-postgres
      key: POSTGRES_PASSWORD
{{- else }}
- name: PGPASSWORD
  value: {{ .Values.postgres.password | quote }}
{{- end }}
{{- end }}
