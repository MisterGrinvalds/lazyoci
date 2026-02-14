package app.authz

default allow := false

# Allow authenticated users to read any resource.
allow if {
    input.method == "GET"
    input.user != ""
}

# Allow admins to perform any action.
allow if {
    input.user in data.admins
}

# Deny requests with missing authentication.
deny if {
    input.user == ""
}
