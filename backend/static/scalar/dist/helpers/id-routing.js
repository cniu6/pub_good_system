import { slugify } from "@scalar/helpers/string/slugify";
//#region src/helpers/id-routing.ts
var sanitizeBasePath = (basePath) => {
	return basePath.replace(/^\/+|\/+$/g, "");
};
var isHashBasePath = (basePath) => basePath.startsWith("#");
var sanitizeHashBasePath = (basePath) => {
	return basePath.replace(/^#+/, "").replace(/\/+$/g, "");
};
var applySlugPrefix = (base, slugPrefix) => {
	return slugPrefix ? `${slugPrefix}${base ? "/" : ""}${base}` : base;
};
var stripBasePathPrefix = (value, basePath) => {
	if (value === basePath) return "";
	if (value.startsWith(`${basePath}/`)) return value.slice(basePath.length + 1);
	return null;
};
/**
* Builds the URL-encoded, slash-prefixed form of a sanitized base path.
*
* Each segment is encoded separately so the result matches how the browser stores
* `location.pathname` (which encodes within segments but leaves the slashes between them).
*/
var encodeBasePath = (sanitized) => {
	if (!sanitized) return "";
	return `/${sanitized.split("/").map((segment) => encodeURIComponent(segment)).join("/")}`;
};
/** Extracts an element id from the hash when using hash routing */
var getIdFromHash = (location, slugPrefix) => {
	const url = typeof location === "string" ? new URL(location) : location;
	return applySlugPrefix(decodeURIComponent(url.hash.slice(1)), slugPrefix);
};
/** Extracts an element id from the path when using path routing */
var getIdFromPath = (location, basePath, slugPrefix) => {
	const url = typeof location === "string" ? new URL(location) : location;
	const basePathWithSlash = encodeBasePath(sanitizeBasePath(basePath));
	if (url.pathname.startsWith(basePathWithSlash)) {
		const remainder = url.pathname.slice(basePathWithSlash.length);
		return applySlugPrefix(decodeURIComponent(remainder.startsWith("/") ? remainder.slice(1) : remainder), slugPrefix);
	}
	return slugPrefix ?? "";
};
/** Extracts an element id from a hash-prefixed basePath */
var getIdFromHashBasePath = (location, basePath, slugPrefix) => {
	const url = typeof location === "string" ? new URL(location) : location;
	const remainder = stripBasePathPrefix(decodeURIComponent(url.hash.slice(1)), sanitizeHashBasePath(basePath));
	if (remainder !== null) return applySlugPrefix(remainder, slugPrefix);
	return slugPrefix ?? "";
};
/** Determines whether a URL matches the provided basePath. */
var matchesBasePath = (location, basePath) => {
	const url = typeof location === "string" ? new URL(location) : location;
	if (isHashBasePath(basePath)) {
		const hash = decodeURIComponent(url.hash);
		return hash === basePath || hash.startsWith(`${basePath}/`);
	}
	const basePathWithSlash = encodeBasePath(sanitizeBasePath(basePath));
	return url.pathname === basePathWithSlash || url.pathname.startsWith(`${basePathWithSlash}/`);
};
/**
* Extracts a navigation id from a URL based on the routing type
*
* @param url - The URL to extract the id from
* @param basePath - The base path used in path routing
* @param slugPrefix - If the document slug is not expected in the URL then we must prefix it
*/
var getIdFromUrl = (url, basePath, slugPrefix) => {
	if (typeof basePath !== "string") return getIdFromHash(url, slugPrefix);
	return isHashBasePath(basePath) ? getIdFromHashBasePath(url, basePath, slugPrefix) : getIdFromPath(url, basePath, slugPrefix);
};
/**
* Strips the first segment from an id and preserves trailing slashes
* Used in single-document mode where the document slug is not needed in the URL
*
* @param id - The full id to process
* @returns The id with the first segment removed, preserving trailing slash if present
*/
var stripFirstSegment = (id) => {
	const hasTrailingSlash = id.endsWith("/");
	const result = id.split("/").filter(Boolean).slice(1).join("/");
	return hasTrailingSlash && result ? `${result}/` : result;
};
/**
* Generate a new URL and applies the ID to the path or hash
* depending on the type of routing used
*
* @param id - The id to apply to the URL
* @param basePath - The base path used in path routing
* @param isMultiDocument - Whether the document is multi-document or single-document. Single-document documents will strip the document slug from the id
*/
var makeUrlFromId = (_id, basePath, isMultiDocument) => {
	if (typeof window === "undefined") return;
	/** When there is only 1 document we don't need to include the document name in the URL */
	const id = isMultiDocument ? _id : stripFirstSegment(_id);
	const url = new URL(window.location.href);
	if (typeof basePath === "string") if (isHashBasePath(basePath)) url.hash = [sanitizeHashBasePath(basePath), id].filter(Boolean).join("/");
	else url.pathname = `${sanitizeBasePath(basePath)}/${id}`;
	else url.hash = id;
	return url;
};
var escapeRegex = (value) => value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
/** Carrier for an id that lives in the bare hash fragment (`#<id>`). */
var bareHashCarrier = (url) => ({
	value: decodeURIComponent(url.hash.slice(1)),
	write: (next) => {
		url.hash = next;
	}
});
/**
* Locates the navigation id inside a URL regardless of the routing mode.
*
* The id can live in three places — the bare hash, a hash base path (`#<base>/<id>`), or the
* pathname after a path base path. Each carrier pairs the decoded id with a `write` callback, so
* callers edit ids in id-space without re-deriving the routing chrome for each mode.
*
* Multiple carriers are returned: the canonical location for the active routing mode comes first,
* followed by a legacy bare-hash carrier. Both path and hash-base-path routing include the bare hash
* so a stale `#default/model/User` bookmark (left over from before the base path was configured) is
* still canonicalized — the previous implementation rewrote such doc-slug-anchored hashes on every
* load regardless of the routing mode. Callers rewrite every carrier that matches, not just the
* first, so a legacy id sitting in both the pathname and the hash is cleaned up in one pass.
*/
var locateIdCarriers = (url, basePath) => {
	if (typeof basePath !== "string") return [bareHashCarrier(url)];
	if (isHashBasePath(basePath)) {
		const base = sanitizeHashBasePath(basePath);
		return [{
			value: stripBasePathPrefix(decodeURIComponent(url.hash.slice(1)), base) ?? "",
			write: (next) => {
				url.hash = [base, next].filter(Boolean).join("/");
			}
		}, bareHashCarrier(url)];
	}
	const base = sanitizeBasePath(basePath);
	const remainder = stripBasePathPrefix(url.pathname, encodeBasePath(base));
	return [{
		value: remainder === null ? "" : decodeURIComponent(remainder),
		write: (next) => {
			url.pathname = base ? `/${base}/${next}` : `/${next}`;
		}
	}, bareHashCarrier(url)];
};
/**
* Applies redirect rules to an id and returns the first rewrite that changes it.
*
* Rules are tried in order and the search stops at the first match, so list more specific rules
* before more general ones. Returns the id unchanged when no rule applies.
*/
var applyIdRedirects = (id, redirects) => {
	for (const { match, replace } of redirects) {
		const next = typeof replace === "string" ? id.replace(match, replace) : id.replace(match, replace);
		if (next !== id) return next;
	}
	return id;
};
/** Optional `(tag-group/<n>/)?tag/<slug>/` chrome that may sit in front of a section segment. */
var TAG_CHROME = "(?:(?:tag-group/[^/]+/)?tag/[^/]+/)?";
/**
* Builds the list of id redirects for the current document.
*
* Add an entry here to support a new redirect. Each rule runs against every URL shape (hash, hash
* base path, path) for free, because rules operate on routing-agnostic ids.
*/
var buildRedirects = ({ modelsSectionSlug, documentSlug, isMultiDocument }) => {
	const escapedDoc = escapeRegex(documentSlug);
	const documentPrefix = isMultiDocument ? `${escapedDoc}/${TAG_CHROME}` : `(?:${escapedDoc}/${TAG_CHROME})?`;
	const renameSectionSegment = (from) => ({
		match: new RegExp(`^(${documentPrefix})${escapeRegex(from)}/`),
		replace: (_match, prefix) => `${prefix}${modelsSectionSlug}/`
	});
	const redirects = [renameSectionSegment("model")];
	if (modelsSectionSlug !== "models") redirects.push(renameSectionSegment("models"));
	return redirects;
};
/**
* Markers that separate an operation id from a schema path within an anchor id.
* `responses` mirrors the request-body/parameter markers so response property
* anchors (e.g. `operation.responses.200.name`) resolve back to the operation id.
* Without it, deep links into responses cannot find their operation.
*/
var SCHEMA_PARAM_MARKERS = [
	".body.",
	".path.",
	".query.",
	".header.",
	".responses."
];
/**
* Builds redirect rules mapping each webhook's legacy dot-dropped slug to its current slug.
*
* Webhook ids used to slugify the event name with no options, which dropped dots
* (`account_holder.created` -> `account-holdercreated`); the current id keeps them
* (`account-holder.created`). That rewrite cannot be reversed generically, so we emit one exact rule
* per webhook whose legacy slug differs from the current one. Each rule matches the invariant
* `webhook/<METHOD>/<legacy-slug>` tail, so it applies regardless of the document/tag prefix in front
* of it and preserves any sub-anchor after it. A sub-anchor is either a `/` segment or a dot-joined
* schema property anchor (e.g. `.body.id`, see {@link SCHEMA_PARAM_MARKERS}); a dot followed by
* anything else is not treated as a boundary, because the legacy slug never contains a dot while a
* current webhook slug can, so such a URL belongs to a different webhook.
*
* Because the legacy slug is lossy, two things can make a redirect unsafe, and both are dropped so an
* ambiguous link falls through to normal not-found handling instead of landing on the wrong webhook:
* - the legacy slug is already another webhook's *current* slug (redirecting it would hijack a valid
*   URL), or
* - two webhooks collapse to the same legacy slug (the old bookmark is genuinely ambiguous).
*/
var buildWebhookRedirects = (webhooks) => {
	const key = (method, slug) => `${method.toUpperCase()}/${slug}`;
	const currentKeys = /* @__PURE__ */ new Set();
	const legacyCounts = /* @__PURE__ */ new Map();
	for (const { name, method, id } of webhooks) {
		currentKeys.add(key(method, id.slice(id.lastIndexOf("/") + 1)));
		const legacySlug = slugify(name);
		if (legacySlug) {
			const legacyKey = key(method, legacySlug);
			legacyCounts.set(legacyKey, (legacyCounts.get(legacyKey) ?? 0) + 1);
		}
	}
	const boundary = `(?=$|/|${SCHEMA_PARAM_MARKERS.map(escapeRegex).join("|")})`;
	const redirects = [];
	for (const { name, method, id } of webhooks) {
		const upperMethod = method.toUpperCase();
		const currentSlug = id.slice(id.lastIndexOf("/") + 1);
		const legacySlug = slugify(name);
		const legacyKey = key(method, legacySlug);
		if (!legacySlug || legacySlug === currentSlug || currentKeys.has(legacyKey) || (legacyCounts.get(legacyKey) ?? 0) > 1) continue;
		redirects.push({
			match: new RegExp(`(^|/)webhook/${escapeRegex(upperMethod)}/${escapeRegex(legacySlug)}${boundary}`),
			replace: (_match, prefix) => `${prefix}webhook/${upperMethod}/${currentSlug}`
		});
	}
	return redirects;
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
var redirectUrl = (url, modelsSectionSlug, documentSlug, isMultiDocument, basePath, webhooks = []) => {
	if (!documentSlug) return null;
	const target = new URL(typeof url === "string" ? url : url.toString());
	const redirects = [...buildRedirects({
		modelsSectionSlug,
		documentSlug,
		isMultiDocument
	}), ...buildWebhookRedirects(webhooks)];
	let didRedirect = false;
	for (const carrier of locateIdCarriers(target, basePath)) {
		const rewritten = applyIdRedirects(carrier.value, redirects);
		if (rewritten !== carrier.value) {
			carrier.write(rewritten);
			didRedirect = true;
		}
	}
	return didRedirect ? target : null;
};
/** Extracts the schema parameters from the id if they are present */
var getSchemaParamsFromId = (id) => {
	let markerIndex = -1;
	for (const marker of SCHEMA_PARAM_MARKERS) {
		const index = id.indexOf(marker);
		if (index !== -1 && (markerIndex === -1 || index < markerIndex)) markerIndex = index;
	}
	if (markerIndex === -1) return {
		rawId: id,
		params: ""
	};
	return {
		rawId: id.slice(0, markerIndex),
		params: id.slice(markerIndex + 1)
	};
};
//#endregion
export { getIdFromHash, getIdFromHashBasePath, getIdFromPath, getIdFromUrl, getSchemaParamsFromId, makeUrlFromId, matchesBasePath, redirectUrl, sanitizeBasePath };

//# sourceMappingURL=id-routing.js.map