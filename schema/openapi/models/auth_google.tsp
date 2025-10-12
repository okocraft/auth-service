namespace AuthAPI.Models.Google {
  @friendlyName("GoogleFirstLoginRequest")
  model GoogleFirstLoginRequest {
    @format("url")
    @doc("the login key")
    login_key: string;
  }

  @friendlyName("GoogleLoginRequest")
  model GoogleLoginRequest {
    @format("url")
    @doc("the url of the page currently being viewed")
    current_url: string;
  }

  @friendlyName("GoogleLoginResponse")
  model GoogleLoginResponse {
    @format("url")
    @doc("the Google's login page URL")
    redirect_url: string;
  }

  @friendlyName("GoogleLoginResult")
  enum GoogleLoginResult {
    Success: "success",
    NotEnabled: "not_enabled",
    InvalidToken: "invalid_token",
    UserNotFound: "user_not_found",
    LoginKeyNotFound: "login_key_not_found",
    AlreadyLinked: "already_linked",
    InternalError: "internal_error",
  }
}
