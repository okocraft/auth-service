import "../../models/auth_google.tsp";
import "@typespec/http";
import "@typespec/openapi3";

using TypeSpec.Http;
using TypeSpec.OpenAPI;
using AuthAPI.Models.Google;

@tag("GoogleAuthAPI")
@route("/google")
namespace AuthAPI.Google.Endpoints {
  @route("/link")
  @post
  @operationId("linkWithGoogle")
  @doc("First login with Google Account")
  op linkWithGoogle(@body _: GoogleFirstLoginRequest): {
    @statusCode
    statusCode: 200;

    @doc("the response for Google login")
    @body
    _: GoogleLoginResponse;
  } | {
    @doc("if Google login is not enabled")
    @statusCode
    statusCode: 404;
  };

  @route("/login")
  @post
  @operationId("loginWithGoogle")
  @doc("Login with Google Account")
  op loginWithGoogle(@body _: GoogleLoginRequest): {
    @statusCode
    statusCode: 200;

    @doc("the response for Google login")
    @body
    _: GoogleLoginResponse;
  } | {
    @doc("if Google login is not enabled")
    @statusCode
    statusCode: 404;
  };

  @route("/callback")
  @get
  @operationId("callbackFromGoogle")
  @doc("Callback from Google")
  op callbackFromGoogle(): {
    @doc("redirect to login result page")
    @statusCode
    statusCode: 307;
  } | {
    @doc("if Google login is not enabled")
    @statusCode
    statusCode: 404;
  };
}
