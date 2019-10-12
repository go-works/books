
import { writable } from 'svelte/store';

export const currentlySelectedIdx = writable(-1);

// currentlySelectedIdx.subscribe(v => {
//     console.log("currentlySelectedIdx:", v);
// })

export function storeSet(key, val) {
    if (window.localStorage) {
        window.localStorage.setItem(key, val);
    }
}

export function storeClear(key) {
    if (window.localStorage) {
        window.localStorage.removeItem(key);
    }
}

export function storeGet(key) {
    if (window.localStorage) {
        return window.localStorage.getItem(key);
    }
    return "";
}

var keyScrollPos = "scrollPos";
var keyIndexView = "indexView";

export function scrollPosSet(pos) {
    storeSet(keyScrollPos, pos);
}

export function scrollPosGet(pos) {
    return storeGet(keyScrollPos);
}

export function scrollPosClear() {
    storeClear(keyScrollPos);
}

export function viewSet(view) {
    storeSet(keyIndexView, view);
}

export function viewGet() {
    return storeGet(keyIndexView);
}

export function viewClear() {
    storeClear(keyIndexView);
}
