import type { AnyApiReferenceConfiguration } from '@scalar/types/api-reference';
export type { ApiReferenceConfiguration } from '@scalar/types/api-reference';
export { default as ApiReference } from './components/ApiReference.vue.js';
export { default as GettingStarted } from './components/GettingStarted.vue.js';
export { SearchButton, SearchModal } from './features/Search';
export { createEmptySpecification } from './helpers/openapi.js';
export { createApiReference } from './standalone/lib/html-api.js';
export type ReferenceProps = {
    configuration?: AnyApiReferenceConfiguration;
};
//# sourceMappingURL=index.d.ts.map