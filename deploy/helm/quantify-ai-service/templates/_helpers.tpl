{{/*
Expand the name of the chart.
*/}}
{{- define "quantify-ai-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "quantify-ai-service.fullname" -}}
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
{{- define "quantify-ai-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "quantify-ai-service.labels" -}}
helm.sh/chart: {{ include "quantify-ai-service.chart" . }}
{{ include "quantify-ai-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: quantify
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "quantify-ai-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "quantify-ai-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: ai-service
{{- end }}

{{/*
Create the name of the service account to use.
*/}}
{{- define "quantify-ai-service.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "quantify-ai-service.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Resolve the container image reference.
*/}}
{{- define "quantify-ai-service.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end }}

{{/*
Name of the ConfigMap holding non-sensitive configuration.
*/}}
{{- define "quantify-ai-service.configMapName" -}}
{{- printf "%s-config" (include "quantify-ai-service.fullname" .) -}}
{{- end }}

{{/*
Name of the Secret holding sensitive configuration.
*/}}
{{- define "quantify-ai-service.secretName" -}}
{{- if .Values.externalSecrets.enabled -}}
{{- default (printf "%s-external" (include "quantify-ai-service.fullname" .)) .Values.externalSecrets.targetName -}}
{{- else -}}
{{- printf "%s-secret" (include "quantify-ai-service.fullname" .) -}}
{{- end -}}
{{- end }}
