{{/*
nginx auth Secret name
*/}}
{{- define "mimir.proxy.nginx.authSecret" -}}
{{ .Values.proxy.nginx.basicAuth.existingSecret | default (include "mimir.resourceName" (dict "ctx" . "component" "proxy-nginx") ) }}
{{- end }}

{{/*
Name of the proxy Service resource
*/}}
{{- define "mimir.proxy.service.name" -}}
{{ .Values.proxy.service.nameOverride | default (include "mimir.resourceName" (dict "ctx" . "component" "proxy") ) }}
{{- end }}
