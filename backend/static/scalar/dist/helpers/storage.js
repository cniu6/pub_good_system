import { REFERENCE_LS_KEYS, safeLocalStorage } from "@scalar/helpers/object/local-storage";
import { AuthSchema } from "@scalar/workspace-store/entities/auth";
import { coerceValue } from "@scalar/workspace-store/schemas/typebox-coerce";
//#region src/helpers/storage.ts
var storage = safeLocalStorage();
/**
* Provides an interface to store and retrieve the selected client value
* in local storage.
*/
var clientStorage = () => {
	const key = REFERENCE_LS_KEYS.SELECTED_CLIENT;
	return {
		/**
		* Gets the stored selected client from local storage.
		*/
		get: () => {
			return storage.getItem(key);
		},
		/**
		* Stores the selected client value in local storage.
		* @param value The value to store
		*/
		set: (value) => {
			storage.setItem(key, value);
		}
	};
};
/**
* Provides an interface to store and retrieve authentication scheme
* information in local storage, including both the available schemes and
* the user's selected schemes.
*/
var authStorage = () => {
	const getKey = (slug) => {
		return `${REFERENCE_LS_KEYS.AUTH}-${slug}`;
	};
	return {
		/**
		* Retrieves and coerces the authentication schemes stored in local storage.
		*/
		getAuth: (slug) => {
			return coerceValue(AuthSchema, JSON.parse(storage.getItem(getKey(slug)) ?? "{}"));
		},
		/**
		* Stores the authentication schemes in local storage.
		* @param value The Auth object to stringify and store.
		*/
		setAuth: (slug, value) => {
			storage.setItem(getKey(slug), JSON.stringify(value));
		}
	};
};
//#endregion
export { authStorage, clientStorage };

//# sourceMappingURL=storage.js.map