template {
    destination = "/opt/talend/secrets/<APPSVC_SECRETS_DESTINATION>"
    contents = <<EOH
    <APPSVC_TEMPLATE_CONTENT>
    EOH
    command = "<APPSVC_TEMPLATE_COMMAND_TO_RUN>"
    wait {
    min = "1s"
    max = "2s"
    }
}