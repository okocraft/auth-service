import "../models/auth.tsp";
import "@typespec/http";
import "@typespec/openapi3";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using AuthAPI.Models;

@tag("AuthAPI")
namespace AuthAPI.Endpoints {
  @route("/logout")
  @post
  @operationId("logout")
  @doc("Invalidate refresh_token and access_token")
  op logout(
    @cookie refresh_token: string,
    @header("X-CSRF-Token") csrfToken?: string,
  ): {
    @doc("this endpoint always returns 204")
    @statusCode
    statusCode: 204;
  };

  @route("/refresh")
  @post
  @operationId("refreshAccessToken")
  @doc("Refresh access token")
  op refreshAccessToken(
    @cookie refresh_token: string,
    @header("X-CSRF-Token") csrfToken?: string,
  ): {
    @doc("successfully refreshed access token")
    @statusCode
    statusCode: 200;

    @body _: AccessTokenResponse;
  } | {
    @doc("invalid refresh token")
    @statusCode
    statusCode: 401;
  };
}
