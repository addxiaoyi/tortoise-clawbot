{{/*
Expand the name of the chart.
*/}}
{{- define "tortoise.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "tortoise.fullname" -}}
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
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tortoise.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "_" "-" | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "tortoise.labels" -}}
helm.sh/chart: {{ include "tortoise.chart" . }}
{{ include "tortoise.fullname" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "tortoise.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tortoise.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
