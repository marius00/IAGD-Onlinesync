Installation

Authentication
See https://blogs.msdn.microsoft.com/appserviceteam/2016/11/17/url-authorization-rules/
1) Enable azure AD
2) Allow unathenticated (json handles it)
3) Make sure the json is saved as UTF-8 without BOM

* AllowAnonymous: Allows anonymous clients to access the resource.
* RedirectToLoginPage: If a user is unauthenticated, they will be redirected to the login page of the default identity provider that was configured in the portal (for example, Azure Active Directory).
* RejectWith401: Unauthenticated requests will fail with an HTTP 401 status.
* RejectWith404: Unauthenticated requests will fail with an HTTP 404 status. You would choose this over RejectWith401 

For the custom binero email setup, the email-secret environmental must be set for authentication.