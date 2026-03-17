action "pfsense_execute_php_command" "example" {
  config {
    command = "print(json_encode(phpversion()));"
  }
}
