import { type ModelsSectionLabel } from '@scalar/types/api-reference';
import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { OpenApiDocument } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { FuseData } from '../../../features/Search/types';
/** Documents the search index can ingest. AsyncAPI is supported for headings, tags, and models; channels/operations/messages are not indexed yet. */
type SearchableDocument = OpenApiDocument | AsyncApiDocument;
type CreateSearchIndexOptions = {
    labels?: SearchIndexLabels;
    modelsSectionLabel?: ModelsSectionLabel;
};
type SearchIndexLabels = {
    heading: string;
    tagGroup: string;
    webhook: string;
    webhooks: string;
    introduction: string;
};
/**
 * Create a search index from a list of entries.
 */
export declare function createSearchIndex(document: SearchableDocument | undefined, options?: CreateSearchIndexOptions): FuseData[];
export {};
//# sourceMappingURL=create-search-index.d.ts.map