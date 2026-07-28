{{/*
Expand the name of the chart.
*/}}
{{- define "quantify-backend.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Truncated at 63 chars because some Kubernetes name fields are limited to this
(by the DNS naming spec). If the release name contains the chart name it will
be used as a full name.
*/}}
{{- define "quantify-backend.fullname" -}}
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
{{- define "quantify-backend.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "quantify-backend.labels" -}}
helm.sh/chart: {{ include "quantify-backend.chart" . }}
{{ include "quantify-backend.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: quantify
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "quantify-backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "quantify-backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Create the name of the service account to use.
*/}}
{{- define "quantify-backend.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "quantify-backend.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Resolve the container image reference.
*/}}
{{- define "quantify-backend.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end }}

{{/*
Name of the ConfigMap holding non-sensitive configuration.
*/}}
{{- define "quantify-backend.configMapName" -}}
{{- printf "%s-config" (include "quantify-backend.fullname" .) -}}
{{- end }}

{{/*
Name of the Secret holding sensitive configuration. When external secrets are
enabled this resolves to the ExternalSecret-managed Secret name.
*/}}
{{- define "quantify-backend.secretName" -}}
{{- if .Values.externalSecrets.enabled -}}
{{- default (printf "%s-external" (include "quantify-backend.fullname" .)) .Values.externalSecrets.targetName -}}
{{- else -}}
{{- printf "%s-secret" (include "quantify-backend.fullname" .) -}}
{{- end -}}
{{- end }}
