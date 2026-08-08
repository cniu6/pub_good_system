export declare const sanitizeBasePath: (basePath: string) => string;
/** Extracts an element id from the hash when using hash routing */
export declare const getIdFromHash: (location: string | URL, slugPrefix: string | undefined) => string;
/** Extracts an element id from the path when using path routing */
export declare const getIdFromPath: (location: string | URL, basePath: string, slugPrefix: string | undefined) => string;
/** Extracts an element id from a hash-prefixed basePath */
export declare const getIdFromHashBasePath: (location: string | URL, basePath: string, slugPrefix: string | undefined) => string;
/** Determines whether a URL matches the provided basePath. */
export declare const matchesBasePath: (location: string | URL, basePath: string) => boolean;
/**
 * Extracts a navigation id from a URL based on the routing type
 *
 * @param url - The URL to extract the id from
 * @param basePath - The base path used in path routing
 * @param slugPrefix - If the document slug is not expected in the URL then we must prefix it
 */
export declare const getIdFromUrl: (url: string | URL, basePath: string | undefined, slugPrefix: string | undefined) => string;
/**
 * Generate a new URL and applies the ID to the path or hash
 * depending on the type of routing used
 *
 * @param id - The id to apply to the URL
 * @param basePath - The base path used in path routing
 * @param isMultiDocument - Whether the document is multi-document or single-document. Single-document documents will strip the document slug from the id
 */
export declare const makeUrlFromId: (_id: string, basePath: string | undefined, isMultiDocument: boolean) => URL | undefined;
/**
 * Source data for a single webhook redirect: the raw event name, its method, and the current id.
 *
 * The current slug is read back off the id, so this layer does not need to know the (dot-keeping)
 * slug rule — it only reproduces the legacy (dot-dropping) slug to match old bookmarks.
 */
export type WebhookRedirectSource = {
    /** Raw webhook event name from the OpenAPI `webhooks` key, e.g. `account_holder.created`. */
    name: string;
    /** HTTP method of the webhook operation. */
    method: string;
    /** Current navigation id of the webhook, e.g. `default/webhook/POST/account-holder.created`. */
    id: string;
};
/**
 * Rewrites navigation ids in a URL to their current form.
 *
 * The URL is reduced to its routing-agnostic id (see {@link locateIdCarriers}), each redirect rule
 * is applied in id-space (see {@link buildRedirects} for models and {@link buildWebhookRedirects} for
 * webhooks), and the result is spliced back into the original location. This handles hash,
 * hash-base-path, and path routing uniformly, so a new redirect only has to be added to the list once.
 *
 * Returns the canonicalized URL when a rewrite happens, or null otherwise.
 */
export declare const redirectUrl: (url: string | URL, modelsSectionSlug: string, documentSlug: string, isMultiDocument: boolean, basePath?: string, webhooks?: WebhookRedirectSource[]) => URL | null;
/** Extracts the schema parameters from the id if they are present */
export declare const getSchemaParamsFromId: (id: string) => {
    rawId: string;
    params: string;
};
//# sourceMappingURL=id-routing.d.ts.map