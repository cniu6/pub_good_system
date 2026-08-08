import { type Auth } from '@scalar/workspace-store/entities/auth';
/**
 * Provides an interface to store and retrieve the selected client value
 * in local storage.
 */
export declare const clientStorage: () => {
    /**
     * Gets the stored selected client from local storage.
     */
    get: () => string | null;
    /**
     * Stores the selected client value in local storage.
     * @param value The value to store
     */
    set: (value: string) => void;
};
/**
 * Provides an interface to store and retrieve authentication scheme
 * information in local storage, including both the available schemes and
 * the user's selected schemes.
 */
export declare const authStorage: () => {
    /**
     * Retrieves and coerces the authentication schemes stored in local storage.
     */
    getAuth: (slug: string) => {
        selected: {
            path?: {
                [x: string]: {
                    [x: string]: {
                        selectedIndex: number;
                        selectedSchemes: {}[];
                    };
                };
            } | undefined;
            document?: {
                selectedIndex: number;
                selectedSchemes: {}[];
            } | undefined;
        };
        secrets: {
            [x: string]: ({
                type: "apiKey";
            } & {
                'x-scalar-secret-token': string;
            }) | ({
                type: "http";
            } & {
                'x-scalar-secret-token': string;
            } & {
                'x-scalar-secret-username': string;
                'x-scalar-secret-password': string;
            }) | ({
                type: "oauth2";
            } & {
                implicit?: ({
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-redirect-uri': string;
                }) | undefined;
                password?: ({
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-username': string;
                    'x-scalar-secret-password': string;
                } & {
                    'x-scalar-secret-client-secret': string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                }) | undefined;
                clientCredentials?: ({
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-client-secret': string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                }) | undefined;
                authorizationCode?: ({
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-client-secret': string;
                } & {
                    'x-scalar-secret-redirect-uri': string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                }) | undefined;
            }) | {
                implicit?: ({
                    scopes: {
                        [x: string]: string;
                    };
                    refreshUrl: string;
                } & {
                    'x-scalar-security-query'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-scalar-security-body'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-tokenName'?: string | undefined;
                } & {
                    'x-scalar-secret-auth-url'?: string | undefined;
                } & {
                    'x-scalar-secret-token-url'?: string | undefined;
                } & {
                    authorizationUrl: string;
                } & {
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-redirect-uri': string;
                }) | undefined;
                password?: ({
                    scopes: {
                        [x: string]: string;
                    };
                    refreshUrl: string;
                } & {
                    'x-scalar-security-query'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-scalar-security-body'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-tokenName'?: string | undefined;
                } & {
                    'x-scalar-secret-auth-url'?: string | undefined;
                } & {
                    'x-scalar-secret-token-url'?: string | undefined;
                } & {
                    tokenUrl: string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                } & {
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-username': string;
                    'x-scalar-secret-password': string;
                } & {
                    'x-scalar-secret-client-secret': string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                }) | undefined;
                clientCredentials?: ({
                    scopes: {
                        [x: string]: string;
                    };
                    refreshUrl: string;
                } & {
                    'x-scalar-security-query'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-scalar-security-body'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-tokenName'?: string | undefined;
                } & {
                    'x-scalar-secret-auth-url'?: string | undefined;
                } & {
                    'x-scalar-secret-token-url'?: string | undefined;
                } & {
                    tokenUrl: string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                } & {
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-client-secret': string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                }) | undefined;
                authorizationCode?: ({
                    scopes: {
                        [x: string]: string;
                    };
                    refreshUrl: string;
                } & {
                    'x-scalar-security-query'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-scalar-security-body'?: {
                        [x: string]: string;
                    } | undefined;
                } & {
                    'x-tokenName'?: string | undefined;
                } & {
                    'x-scalar-secret-auth-url'?: string | undefined;
                } & {
                    'x-scalar-secret-token-url'?: string | undefined;
                } & {
                    authorizationUrl: string;
                    tokenUrl: string;
                } & {
                    'x-usePkce': "plain" | "SHA-256" | "no";
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                } & {
                    'x-scalar-secret-client-id': string;
                } & {
                    'x-scalar-secret-token': string;
                } & {
                    'x-scalar-secret-refresh-token'?: string | undefined;
                } & {
                    'x-scalar-secret-client-secret': string;
                } & {
                    'x-scalar-secret-redirect-uri': string;
                } & {
                    'x-scalar-credentials-location'?: "body" | "header" | undefined;
                }) | undefined;
                type: "openIdConnect";
            } | ({
                type: "userPassword" | "plain" | "scramSha256" | "scramSha512";
            } & {
                'x-scalar-secret-username': string;
                'x-scalar-secret-password': string;
            }) | ({
                type: "X509";
            } & {
                'x-scalar-secret-client-certificate': string;
            } & {
                'x-scalar-secret-private-key': string;
            }) | ({
                type: "symmetricEncryption" | "asymmetricEncryption";
            } & {
                'x-scalar-secret-token': string;
            }) | ({
                type: "gssapi";
            } & {
                'x-scalar-secret-service-name': string;
            });
        };
    };
    /**
     * Stores the authentication schemes in local storage.
     * @param value The Auth object to stringify and store.
     */
    setAuth: (slug: string, value: Auth) => void;
};
//# sourceMappingURL=storage.d.ts.map