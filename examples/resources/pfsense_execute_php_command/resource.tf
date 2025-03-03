# get pfsense php version
resource "pfsense_execute_php_command" "php_version" {
  command = "print(json_encode(phpversion()));"
}

output "php_version" {
  value = pfsense_execute_php_command.php_version.result
}

# create/delete file
resource "pfsense_execute_php_command" "file" {
  command         = <<-EOT
  $file = fopen("/tmp/terraform", "w") or die("unable to open file");
  $currentTime = time();
  fwrite($file, $currentTime);
  fclose($file);
  print(json_encode($currentTime));
  EOT
  destroy_command = <<-EOT
  unlink("/tmp/terraform");
  print(json_encode("file removed"));
  EOT
}
