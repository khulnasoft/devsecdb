{{/*
Allow the release namespace to be overridden for multi-namespace deployments in combined charts
*/}}
{{- define "devsecdb.namespace" -}}
  {{- if .Values.namespaceOverride -}}
    {{- .Values.namespaceOverride -}}
  {{- else -}}
    {{- .Release.Namespace -}}
  {{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "devsecdb.labels" -}}
{{ include "devsecdb.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.devsecdb.version}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "devsecdb.selectorLabels" -}}
app: devsecdb
{{- end }}