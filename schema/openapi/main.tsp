import "./endpoints/auth/auth.tsp";
import "./endpoints/auth/oauth/oauth.tsp";
import "./endpoints/auth/oauth/google/google.tsp";
import "./models/auth.tsp";
import "./models/auth_google.tsp";
import "@typespec/openapi";
import "@typespec/openapi3";
import "@typespec/versioning";

using TypeSpec.Versioning;

@service(#{ title: "Auth API" })
@versioned(Versions)
namespace AuthAPI;

enum Versions {
  `v1.0`,
}
