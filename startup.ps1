# Add environment variable
[Environment]::SetEnvironmentVariable( `
  "CUSTOM_VARIABLE", "", `
  [System.EnvironmentVariableTarget]::Machine)
