{{/*
Expand the name of the chart.
*/}}
{{- define "csi-gcs.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "csi-gcs.fullname" -}}
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
{{- define "csi-gcs.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "csi-gcs.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "csi-gcs.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "csi-gcs" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the priority class to use
*/}}
{{- define "csi-gcs.priorityClassName" -}}
{{- if .Values.priorityClass.create }}
{{- default (include "csi-gcs.fullname" .) .Values.priorityClass.name }}
{{- else }}
{{- default "csi-gcs-priority" .Values.priorityClass.name }}
{{- end }}
{{- end }}
