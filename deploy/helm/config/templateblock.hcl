template {
    destination = "/opt/talend/secrets/<VSI_SECRETS_DESTINATION>"
    contents = <<EOH
    <VSI_SECRETS_TEMPLATE_CONTENT>
    EOH
    command = "<VSI_SECRETS_TEMPLATE_COMMAND_TO_RUN>"
    wait {
        min = "1s"
        max = "2s"
    }
}