import type { StandardSchemaV1 } from "@standard-schema/spec";
import { base64url, createLocalJWKSet, errors, type JSONWebKeySet, jwtVerify } from "jose";

/**
 * The well-known information for an OAuth 2.0 authorization server.
 * @internal
 */
export interface WellKnown {
  /**
   * The URI to the JWKS endpoint.
   */
  jwks_uri: string;
  /**
   * The URI to the token endpoint.
   */
  token_endpoint: string;
  /**
   * The URI to the authorization endpoint.
   */
  authorization_endpoint: string;
}

interface ResponseLike {
  json(): Promise<unknown>;
  ok: Response["ok"];
}
type FetchLike = (...args: any[]) => Promise<ResponseLike>;

/**
 * Configure the client.
 */
export interface ClientConfig {
  /**
   * The client ID. This is just a string to identify your app.
   *
   * If you have a web app and a mobile app, you want to use different client IDs both.
   *
   * @example
   * ```ts
   * {
   *   clientID: "my-client"
   * }
   * ```
   */
  clientID: string;

  /**
   * The URL of your OpenAuth server.
   *
   * @example
   * ```ts
   * {
   *   issuer: "https://auth.myserver.com"
   * }
   * ```
   */
  issuer?: string;

  /**
   * Optionally, override the internally used fetch function.
   *
   * This is useful if you are using a polyfilled fetch function in your application and you
   * want the client to use it too.
   */
  fetch?: FetchLike;
}

export interface AuthorizeOptions {
  /**
   * Enable the PKCE flow. This is for SPA apps.
   *
   * ```ts
   * {
   *   pkce: true
   * }
   * ```
   *
   * @default false
   */
  pkce?: boolean;
}

/**
 * The challenge that you can use to verify the code.
 */
export type Challenge = {
  /**
   * The state that was sent to the redirect URI.
   */
  state: string;
  /**
   * The verifier that was sent to the redirect URI.
   */
  verifier?: string;
};

export interface VerifyOptions {
  /**
   * Optionally, pass in the refresh token.
   *
   * If passed in, this will automatically refresh the access token if it has expired.
   */
  refresh?: string;
  /**
   * @internal
   */
  issuer?: string;
  /**
   * @internal
   */
  audience?: string;
  /**
   * Optionally, override the internally used fetch function.
   *
   * This is useful if you are using a polyfilled fetch function in your application and you
   * want the client to use it too.
   */
  fetch?: FetchLike;
}

/**
 * The tokens returned by the auth server.
 */
export interface Tokens {
  /**
   * The access token.
   */
  access: string;
  /**
   * The refresh token.
   */
  refresh: string;

  /**
   * The number of seconds until the access token expires.
   */
  expiresIn: number;
}

export interface VerifyResult<T extends SubjectSchema> {
  /**
   * This is always `undefined` when the verify is successful.
   */
  err?: undefined;
  /**
   * Returns the refreshed tokens only if they’ve been refreshed.
   *
   * If they are still valid, this will be undefined.
   */
  tokens?: Tokens;
  /**
   * @internal
   */
  aud: string;
  /**
   * The decoded subjects from the access token.
   *
   * Has the same shape as the subjects you defined when creating the issuer.
   */
  subject: {
    [type in keyof T]: { type: type; properties: StandardSchemaV1.InferOutput<T[type]> };
  }[keyof T];
}

/**
 * Returned when the verify call fails.
 */
export interface VerifyError {
  /**
   * The type of error that occurred. You can handle this by checking the type.
   *
   * @example
   * ```ts
   * import { InvalidRefreshTokenError } from "@openauthjs/openauth/error"
   *
   * console.log(err instanceof InvalidRefreshTokenError)
   *```
   */
  err: InvalidRefreshTokenError | InvalidAccessTokenError;
}

/**
 * Returned when the refresh is successful.
 */
export interface RefreshSuccess {
  /**
   * This is always `false` when the refresh is successful.
   */
  err: false;
  /**
   * Returns the refreshed tokens only if they've been refreshed.
   *
   * If they are still valid, this will be `undefined`.
   */
  tokens?: Tokens;
}

/**
 * Returned when the refresh fails.
 */
export interface RefreshError {
  /**
   * The type of error that occurred. You can handle this by checking the type.
   *
   * @example
   * ```ts
   * import { InvalidRefreshTokenError } from "@openauthjs/openauth/error"
   *
   * console.log(err instanceof InvalidRefreshTokenError)
   *```
   */
  err: InvalidRefreshTokenError | InvalidAccessTokenError;
}

export function createClient(cfg: ClientConfig) {
  const jwksCache = new Map<string, ReturnType<typeof createLocalJWKSet>>();
  const issuerCache = new Map<string, WellKnown>();

  const issuer = cfg.issuer || process.env.OPENAUTH_ISSUER;
  if (!issuer) {
    throw new Error("No issuer");
  }
  const f = cfg.fetch ?? fetch;

  async function getIssuer() {
    const cached = issuerCache.get(issuer!);
    if (cached) {
      return cached;
    }

    const wellKnown: WellKnown = await f(`${issuer}/.well-known/oauth-authorization-server`).then(
      (r) => r.json(),
    );
    issuerCache.set(issuer!, wellKnown);
    return wellKnown;
  }

  async function getJWKS() {
    const wk = await getIssuer();
    const cached = jwksCache.get(issuer!);
    if (cached) return cached;
    const keyset = (await (f || fetch)(wk.jwks_uri).then((r) => r.json())) as JSONWebKeySet;
    const result = createLocalJWKSet(keyset);
    jwksCache.set(issuer!, result);
    return result;
  }

  const self = {
    async authorize(redirectURI: string, response: "code") {
      const result = new URL(`${issuer}/authorize`);
      const challenge: Challenge = {
        state: crypto.randomUUID(),
      };

      result.searchParams.set("client_id", cfg.clientID);
      result.searchParams.set("redirect_uri", redirectURI);
      result.searchParams.set("response_type", response);
      result.searchParams.set("state", challenge.state);
      const pkce = await generatePKCE();
      result.searchParams.set("code_challenge_method", "S256");
      result.searchParams.set("code_challenge", pkce.challenge);
      challenge.verifier = pkce.verifier;
      return {
        challenge,
        url: result.toString(),
      };
    },

    async exchange(code: string, redirectURI: string, verifier?: string) {
      const tokens = await f(`${issuer}/token`, {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: new URLSearchParams({
          grant_type: "authorization_code",
          client_id: cfg.clientID,
          code,
          code_verifier: verifier || "",
          redirect_uri: redirectURI,
        }).toString(),
      });

      const json = (await tokens.json()) as any;
      if (!tokens.ok) {
        return {
          err: new InvalidAuthorizationCodeError(),
        };
      }
      return {
        err: false,
        tokens: {
          access: json.access_token as string,
          refresh: json.refresh_token as string,
          expiresIn: json.expires_in as number,
        },
      };
    },

    async refresh(refresh: string): Promise<RefreshSuccess | RefreshError> {
      const tokens = await f(`${issuer}/token`, {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: new URLSearchParams({
          grant_type: "refresh_token",
          refresh_token: refresh,
        }).toString(),
      });
      const json = (await tokens.json()) as any;
      if (!tokens.ok) {
        return {
          err: new InvalidRefreshTokenError(),
        };
      }
      return {
        err: false,
        tokens: {
          access: json.access_token as string,
          refresh: json.refresh_token as string,
          expiresIn: json.expires_in as number,
        },
      };
    },

    async verify<T extends SubjectSchema>(
      subjects: T,
      token: string,
      options?: VerifyOptions,
    ): Promise<VerifyResult<T> | VerifyError> {
      const jwks = await getJWKS();

      try {
        const result = await jwtVerify<{
          sub: string;
          roles: string[];
        }>(token, jwks, { issuer });

        const type = "user" as keyof T;
        const properties = {
          sub: result.payload.sub,
          roles: result.payload.roles || [],
        };

        const validated = await subjects[type]["~standard"].validate(properties);
        if (!validated.issues) {
          return {
            aud: result.payload.aud as string,
            subject: {
              type: type,
              properties: validated.value,
            } as any,
          };
        }
        return {
          err: new InvalidSubjectError(),
        };
      } catch (e) {
        if (e instanceof errors.JWTExpired && options?.refresh) {
          const refreshed = await this.refresh(options.refresh);
          if (refreshed.err) {
            return refreshed;
          }

          const verified = await self.verify(subjects, refreshed.tokens!.access, {
            refresh: refreshed.tokens!.refresh,
            issuer,
            fetch: options?.fetch,
          });
          if (verified.err) {
            return verified;
          }

          verified.tokens = refreshed.tokens;
          return verified;
        }

        return {
          err: new InvalidAccessTokenError(),
        };
      }
    },
  };
  return self;
}

/**
 * The given refresh token is invalid.
 */
export class InvalidRefreshTokenError extends Error {
  constructor() {
    super("Invalid refresh token");
  }
}

/**
 * The given access token is invalid.
 */
export class InvalidAccessTokenError extends Error {
  constructor() {
    super("Invalid access token");
  }
}

export type Prettify<T> = {
  [K in keyof T]: T[K];
};

/**
 * The given subject is invalid.
 */
export class InvalidSubjectError extends Error {
  constructor() {
    super("Invalid subject");
  }
}

/**
 * The given authorization code is invalid.
 */
export class InvalidAuthorizationCodeError extends Error {
  constructor() {
    super("Invalid authorization code");
  }
}

/**
 * Subject schema is a map of types that are used to define the subjects.
 */
export type SubjectSchema = Record<string, StandardSchemaV1>;

/** @internal */
export type SubjectPayload<T extends SubjectSchema> = Prettify<
  {
    [type in keyof T & string]: {
      type: type;
      properties: StandardSchemaV1.InferOutput<T[type]>;
    };
  }[keyof T & string]
>;

export function createSubjects<Schema extends SubjectSchema = {}>(types: Schema): Schema {
  return { ...types };
}

function generateVerifier(length: number): string {
  const buffer = new Uint8Array(length);
  crypto.getRandomValues(buffer);
  return base64url.encode(buffer);
}

async function generateChallenge(verifier: string, method: "S256" | "plain") {
  if (method === "plain") return verifier;
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const hash = await crypto.subtle.digest("SHA-256", data);
  return base64url.encode(new Uint8Array(hash));
}

async function generatePKCE(length: number = 64) {
  if (length < 43 || length > 128) {
    throw new Error("Code verifier length must be between 43 and 128 characters");
  }
  const verifier = generateVerifier(length);
  const challenge = await generateChallenge(verifier, "S256");
  return {
    verifier,
    challenge,
    method: "S256",
  };
}
