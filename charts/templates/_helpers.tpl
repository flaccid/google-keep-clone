{{- define "google-keep-clone.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "google-keep-clone.namespace" -}}
{{- .Values.global.namespace | default .Release.Namespace -}}
{{- end }}

{{- define "google-keep-clone.backend.selectorLabels" -}}
app: backend
{{- end }}

{{- define "google-keep-clone.frontend.selectorLabels" -}}
app: frontend
{{- end }}

{{- define "google-keep-clone.postgres.selectorLabels" -}}
app: postgres
{{- end }}

{{- define "google-keep-clone.oauth2-proxy.selectorLabels" -}}
app: oauth2-proxy
{{- end }}
