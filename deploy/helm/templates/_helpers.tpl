{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
Truncate at 63 chars characters due to limitations of the DNS system.
*/}}
{{- define "talend-vault-sidecar-injector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "talend-vault-sidecar-injector.fullname" -}}
{{- $name := (include "talend-vault-sidecar-injector.name" .) -}}
{{- printf "%s-%s" .Release.Name $name | trunc 40 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default chart name including the version number
*/}}
{{- define "talend-vault-sidecar-injector.chart" -}}
{{- $name := (include "talend-vault-sidecar-injector.name" .) -}}
{{- printf "%s-%s" $name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{/*
Define mutating webhook failure policy (https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#failure-policy)
Force 'Ignore' if only one replica (because 'Fail' will prevent any pod to start if the only one Vault Sidecar Injector pod is down...)
*/}}
{{- define "talend-vault-sidecar-injector.failurePolicy" -}}
{{- if eq .replicaCount 1.0 -}}
Ignore
{{else}}
{{- .mutatingwebhook.failurePolicy -}}
{{- end -}}
{{- end -}}

{{/*
Define mutating webhook namespace selector
*/}}
{{- define "talend-vault-sidecar-injector.namespaceSelector" -}}
{{- if and (eq .Values.mutatingwebhook.namespaceSelector.boolean true) (eq .Values.mutatingwebhook.namespaceSelector.namespaced false) -}}
namespaceSelector:
  matchLabels:
    vault-injection: enabled
{{- end -}}
{{- if and (eq .Values.mutatingwebhook.namespaceSelector.namespaced true) (eq .Values.mutatingwebhook.namespaceSelector.boolean false) -}}
namespaceSelector:
  matchLabels:
    vault-injection: {{ .Release.Namespace }}
{{- end -}}
{{- if and (eq .Values.mutatingwebhook.namespaceSelector.namespaced true) (eq .Values.mutatingwebhook.namespaceSelector.boolean true) -}}
{{ fail "Cannot enable both mutatingwebhook.namespaceSelector.namespaced and mutatingwebhook.namespaceSelector.boolean values" }}
{{- end -}}
{{- end -}}

{{/*
Define labels which are used throughout the chart files
*/}}
{{- define "talend-vault-sidecar-injector.labels" -}}
com.talend.application: {{ .Values.image.applicationNameLabel }}
com.talend.service: {{ .Values.image.serviceNameLabel }}
chart: {{ include "talend-vault-sidecar-injector.chart" . }}
helm.sh/chart: {{ include "talend-vault-sidecar-injector.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
{{- end -}}

{{/*
Define the docker image (image.path:image.tag).
*/}}
{{- define "talend-vault-sidecar-injector.image" -}}
{{- printf "%s%s:%s" (default "" .imageRegistry) .image.path (default "latest" .image.tag) -}}
{{- end -}}

{{/*
Define the docker image for Job Babysitter sidecar container (image.path:image.tag).
*/}}
{{- define "talend-vault-sidecar-injector.injectconfig.jobbabysitter.image" -}}
{{- printf "%s%s:%s" (default "" .imageRegistry) .injectconfig.jobbabysitter.image.path (default "latest" .injectconfig.jobbabysitter.image.tag) -}}
{{- end -}}

{{/*
Define the docker image for Vault sidecar container (image.path:image.tag).
*/}}
{{- define "talend-vault-sidecar-injector.injectconfig.vault.image" -}}
{{- printf "%s%s:%s" (default "" .imageRegistry) .injectconfig.vault.image.path (default "latest" .injectconfig.vault.image.tag) -}}
{{- end -}}

{{/*
Returns the service name which is by default fixed (not depending on release).
It can be prefixed by the release if the service.prefixWithHelmRelease is true
*/}}
{{- define "talend-vault-sidecar-injector.service.name" -}}
{{- if eq .Values.service.prefixWithHelmRelease true -}}
    {{- $name := .Values.service.name | trunc 63 | trimSuffix "-" -}}
    {{- printf "%s-%s" .Release.Name $name -}}
{{else}}
    {{- .Values.service.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Add Vault flag to skip verification of TLS certificates
*/}}
{{- define "talend-vault-sidecar-injector.vault.cert.skip.verify" -}}
{{- if eq .vault.ssl.verify false -}}
-tls-skip-verify
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for MutatingWebhookConfiguration
*/}}
{{- define "mutatingwebhookconfiguration.apiversion" -}}
{{- if semverCompare ">=1.16" .Capabilities.KubeVersion.Version -}}
"admissionregistration.k8s.io/v1"
{{- else -}}
"admissionregistration.k8s.io/v1beta1"
{{- end -}}
{{- end -}}
