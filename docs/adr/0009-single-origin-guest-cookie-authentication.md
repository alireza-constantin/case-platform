# Single-origin guest cookie authentication

M0 exposes the web application and API on one public origin. An opaque persistent guest-session cookie identifies the guest; in deployed HTTPS it uses HttpOnly, Secure, and SameSite=Lax or a stricter setting compatible with normal navigation. Playthrough IDs and URLs are identifiers, not bearer credentials. The server checks guest ownership before accessing a playthrough.

Mutation endpoints use non-GET methods and JSON requests, with SameSite cookies and Origin validation rejecting clearly cross-origin mutation requests. M0 does not add rotating CSRF tokens, token synchronization, elaborate origin infrastructure, or cross-origin authentication. Revisit protections if deployment requirements change. Accounts are deferred.
