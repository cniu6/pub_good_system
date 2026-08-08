import { getRequiredSecurity } from "../../../../features/Operation/helpers/get-required-security.js";
import { getAsyncApiSecurityRequirements } from "@scalar/workspace-store/channel-example";
//#region src/components/Content/AsyncApi/helpers/get-async-api-required-security.ts
/**
* Build the required-security model for an AsyncAPI operation so the shared
* `OperationScopes` section can render its OAuth / OpenID Connect scopes.
*
* AsyncAPI declares security on the operation (and its traits) as a list of scheme
* references carrying scopes. `getAsyncApiSecurityRequirements` normalises that into the
* same OR-alternative shape the OpenAPI path uses, so we hand it to `getRequiredSecurity`
* and reuse the exact grouping and de-duplication logic.
*
* Only the scopes are needed for this section, so the scheme objects are left unresolved
* (an empty component set). Server-level security is intentionally excluded: it belongs to
* the channel connection, not a single operation.
*/
var getAsyncApiRequiredSecurity = (document, operation) => getRequiredSecurity({ security: getAsyncApiSecurityRequirements(document, operation, null) }, { components: void 0 });
//#endregion
export { getAsyncApiRequiredSecurity };

//# sourceMappingURL=get-async-api-required-security.js.map