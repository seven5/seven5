#/bin/bash
mockgen --package=mock github.com/seven5/seven5 CookieMapper > mock/cookie_mapper.go
mockgen --package=mock github.com/seven5/seven5 OauthCred > mock/oauth_cred.go
mockgen --package=mock github.com/seven5/seven5 DeploymentEnvironment > mock/deployment_environment.go
mockgen --package=mock github.com/seven5/seven5 PageMapper > mock/page_mapper.go
mockgen --package=mock github.com/seven5/seven5 SessionManager > mock/session_manager.go
