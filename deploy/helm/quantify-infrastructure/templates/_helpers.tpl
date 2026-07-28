{{/*
Common labels for infrastructure-level resources.
*/}}
{{- define "quantify-infrastructure.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: quantify
{{- end }}

{{/*
Renders the `from:` list entries that allow the Quantify application
namespaces (and any operator-supplied extraFrom rules) to reach the data layer.
Callers pipe the result through `nindent`.
*/}}
{{- define "quantify-infrastructure.appFrom" -}}
{{- range .Values.networkPolicy.appNamespaces }}
- namespaceSelector:
    matchLabels:
      kubernetes.io/metadata.name: {{ . }}
{{- with $.Values.networkPolicy.appPodLabels }}
  podSelector:
    matchLabels:
      {{- toYaml . | nindent 6 }}
{{- end }}
{{- end }}
{{- with .Values.networkPolicy.extraFrom }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}
