import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { isNonOptionalSecurityRequirement } from "@scalar/workspace-store/helpers/is-non-optional-security-requirement";
//#region src/features/Operation/helpers/get-required-security.ts
/**
* Determine whether an operation requires authentication, using `operation.security ?? document.security`
* as the source of truth. Operation-level `security` fully overrides document-level — including `security: []`,
* which explicitly opts the operation out of auth.
*
* OpenAPI encodes "auth is optional" by including an empty requirement object `{}`. Whenever `{}`
* appears — whether alongside real requirements or as the only entry — auth is treated as optional (state: 'optional').
*/
var getRequiredSecurity = (operation, document) => {
	const securityList = operation?.security ?? document.security ?? [];
	const definedSchemes = document.components?.securitySchemes ?? {};
	let hasEmpty = false;
	const groups = [];
	for (const requirement of securityList) {
		if (!isNonOptionalSecurityRequirement(requirement)) {
			hasEmpty = true;
			continue;
		}
		const schemes = Object.entries(requirement).map(([name, scopes]) => ({
			name,
			scheme: getResolvedRef(definedSchemes[name]),
			scopes: scopes.filter((s) => s.length > 0)
		}));
		if (schemes.length > 0) groups.push({ schemes });
	}
	if (groups.length === 0) return {
		state: hasEmpty ? "optional" : "none",
		requirements: []
	};
	return {
		state: hasEmpty ? "optional" : "required",
		requirements: groups
	};
};
/**
* Required OAuth scopes grouped by security alternative (OR).
*
* Each returned array is the de-duplicated set of scopes for one alternative, kept in
* declaration order. Scopes are only collected for OAuth2 / OpenID Connect schemes, so
* alternatives without any scopes (for example API-key-only auth) are omitted and the
* result is empty when the operation requires no OAuth scopes.
*
* Grouping is preserved intentionally: unioning scopes across alternatives would imply
* that mutually exclusive scope sets are all required at once, which misstates OpenAPI
* OR semantics.
*/
var getRequiredScopeGroups = (requiredSecurity) => {
	const groups = [];
	for (const group of requiredSecurity.requirements) {
		const collected = /* @__PURE__ */ new Set();
		for (const scheme of group.schemes) for (const scope of scheme.scopes) collected.add(scope);
		if (collected.size > 0) groups.push([...collected]);
	}
	return groups;
};
//#endregion
export { getRequiredScopeGroups, getRequiredSecurity };

//# sourceMappingURL=get-required-security.js.map