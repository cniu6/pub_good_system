import { type ModelsSectionLabel } from '@scalar/types/api-reference';
import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { OpenApiDocument } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import { type MaybeRefOrGetter } from 'vue';
import type { FuseData } from '../types';
/**
 * Creates the search index from an OpenAPI or AsyncAPI document.
 */
export declare function useSearchIndex(document: MaybeRefOrGetter<OpenApiDocument | AsyncApiDocument | undefined>, modelsSectionLabel?: MaybeRefOrGetter<ModelsSectionLabel | undefined>, labels?: MaybeRefOrGetter<{
    heading: string;
    tagGroup: string;
    webhook: string;
    webhooks: string;
    introduction: string;
}>): {
    results: import("vue").ComputedRef<{
        item: FuseData;
        refIndex: number;
    }[]>;
    query: import("vue").Ref<string, string>;
};
//# sourceMappingURL=useSearchIndex.d.ts.map