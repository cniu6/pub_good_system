import type { AsyncApiDocument } from '@scalar/types/asyncapi/3.1';
import type { TraversedEntry } from '@scalar/workspace-store/schemas/navigation';
/** The currently selected sidebar filters. Empty / sentinel values disable a filter. */
type AsyncApiNavigationFilter = {
    /** Selected protocol id, or {@link ALL} / undefined for no protocol filter. */
    protocol?: string;
    /** Selected server name, or {@link ALL} / undefined for no server filter. */
    server?: string;
};
/**
 * Filters the AsyncAPI sidebar tree by the selected protocol and/or server.
 *
 * Returns the original entries untouched when no filter is active, so OpenAPI
 * documents and the unfiltered AsyncAPI case pay no cost.
 */
export declare const filterAsyncApiNavigation: (entries: TraversedEntry[], document: AsyncApiDocument, filter: AsyncApiNavigationFilter) => TraversedEntry[];
export {};
//# sourceMappingURL=filter-async-api-navigation.d.ts.map