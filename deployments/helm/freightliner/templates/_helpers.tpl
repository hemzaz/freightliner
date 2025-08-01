{{/*
Expand the name of the chart.
*/}}
{{- define "freightliner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "freightliner.fullname" -}}
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
{{- define "freightliner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "freightliner.labels" -}}
helm.sh/chart: {{ include "freightliner.chart" . }}
{{ include "freightliner.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: replication
{{- end }}

{{/*
Selector labels
*/}}
{{- define "freightliner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "freightliner.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "freightliner.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "freightliner.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the configmap
*/}}
{{- define "freightliner.configMapName" -}}
{{- printf "%s-config" (include "freightliner.fullname" .) }}
{{- end }}

{{/*
Create the name of the secret
*/}}
{{- define "freightliner.secretName" -}}
{{- printf "%s-secrets" (include "freightliner.fullname" .) }}
{{- end }}

{{/*
Create the docker image name
*/}}
{{- define "freightliner.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry -}}
{{- $repository := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end }}

{{/*
Create environment variables from config
*/}}
{{- define "freightliner.envVars" -}}
- name: LOG_LEVEL
  value: {{ .Values.config.logLevel | quote }}
- name: PORT
  value: {{ .Values.config.port | quote }}
- name: AWS_REGION
  value: {{ .Values.config.aws.region | quote }}
- name: GCP_PROJECT_ID
  value: {{ .Values.config.gcp.projectId | quote }}
- name: GCP_REGION
  value: {{ .Values.config.gcp.region | quote }}
- name: WORKER_POOL_SIZE
  value: {{ .Values.config.workerPoolSize | quote }}
- name: MAX_CONCURRENT_REPLICATIONS
  value: {{ .Values.config.maxConcurrentReplications | quote }}
- name: HTTP_TIMEOUT
  value: {{ .Values.config.httpTimeout | quote }}
- name: RETRY_ATTEMPTS
  value: {{ .Values.config.retryAttempts | quote }}
- name: METRICS_ENABLED
  value: {{ .Values.config.metricsEnabled | quote }}
- name: METRICS_PORT
  value: {{ .Values.config.metricsPort | quote }}
- name: ENVIRONMENT
  value: {{ .Values.environment | quote }}
{{- if .Values.config.aws.ecrEndpoint }}
- name: AWS_ECR_ENDPOINT
  value: {{ .Values.config.aws.ecrEndpoint | quote }}
{{- end }}
{{- end }}

{{/*
Create secret environment variables
*/}}
{{- define "freightliner.secretEnvVars" -}}
{{- if .Values.secrets.awsAccessKeyId }}
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: {{ include "freightliner.secretName" . }}
      key: aws-access-key-id
{{- end }}
{{- if .Values.secrets.awsSecretAccessKey }}
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "freightliner.secretName" . }}
      key: aws-secret-access-key
{{- end }}
{{- if .Values.secrets.gcpServiceAccountKey }}
- name: GOOGLE_APPLICATION_CREDENTIALS
  value: /etc/gcp-sa/service-account-key.json
{{- end }}
{{- end }}

{{/*
Create volume mounts
*/}}
{{- define "freightliner.volumeMounts" -}}
- name: config
  mountPath: /etc/freightliner
  readOnly: true
- name: tmp
  mountPath: /tmp
{{- if .Values.secrets.gcpServiceAccountKey }}
- name: gcp-sa
  mountPath: /etc/gcp-sa
  readOnly: true
{{- end }}
{{- if .Values.persistence.enabled }}
- name: data
  mountPath: /data
{{- end }}
{{- end }}

{{/*
Create volumes
*/}}
{{- define "freightliner.volumes" -}}
- name: config
  configMap:
    name: {{ include "freightliner.configMapName" . }}
- name: tmp
  emptyDir: {}
{{- if .Values.secrets.gcpServiceAccountKey }}
- name: gcp-sa
  secret:
    secretName: {{ include "freightliner.secretName" . }}
    items:
      - key: gcp-service-account-key
        path: service-account-key.json
{{- end }}
{{- if .Values.persistence.enabled }}
- name: data
  persistentVolumeClaim:
    claimName: {{ include "freightliner.fullname" . }}-data
{{- end }}
{{- end }}

{{/*
Resource limits and requests
*/}}
{{- define "freightliner.resources" -}}
{{- if .Values.resources }}
resources:
  {{- if .Values.resources.limits }}
  limits:
    {{- if .Values.resources.limits.cpu }}
    cpu: {{ .Values.resources.limits.cpu }}
    {{- end }}
    {{- if .Values.resources.limits.memory }}
    memory: {{ .Values.resources.limits.memory }}
    {{- end }}
  {{- end }}
  {{- if .Values.resources.requests }}
  requests:
    {{- if .Values.resources.requests.cpu }}
    cpu: {{ .Values.resources.requests.cpu }}
    {{- end }}
    {{- if .Values.resources.requests.memory }}
    memory: {{ .Values.resources.requests.memory }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Pod anti-affinity rules
*/}}
{{- define "freightliner.affinity" -}}
{{- if .Values.affinity }}
affinity:
  {{- with .Values.affinity.podAntiAffinity }}
  podAntiAffinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.affinity.podAffinity }}
  podAffinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.affinity.nodeAffinity }}
  nodeAffinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}