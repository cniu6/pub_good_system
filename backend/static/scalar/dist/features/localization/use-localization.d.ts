import type { ApiReferenceLocale, ApiReferenceLocalization, ApiReferenceTextDirection, ApiReferenceTranslationKey, ApiReferenceTranslations } from '@scalar/types/api-reference';
import { type ComputedRef, type MaybeRefOrGetter } from 'vue';
type LocalizationContext = {
    locale: ComputedRef<ApiReferenceLocale>;
    direction: ComputedRef<ApiReferenceTextDirection>;
    translations: ComputedRef<ApiReferenceTranslations>;
    translate: (key: ApiReferenceTranslationKey, params?: Record<string, number | string>) => string;
};
type ResolvedLocalization = {
    locale: ApiReferenceLocale;
    direction: ApiReferenceTextDirection;
    translations: ApiReferenceTranslations;
};
export declare const resolveLocalization: (localization?: ApiReferenceLocalization) => ResolvedLocalization;
export declare const provideLocalization: (localization: MaybeRefOrGetter<ApiReferenceLocalization | undefined>) => LocalizationContext;
export declare const useLocalization: () => LocalizationContext;
export {};
//# sourceMappingURL=use-localization.d.ts.map