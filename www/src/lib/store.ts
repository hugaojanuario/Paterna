interface User {
    id: string;
    email: string;
    token: string;
}

const USER_KEY = "paterna_user";
const SELECTED_KEY = "paterna_selected";

let user: User | null = loadUser();

function loadUser(): User | null {
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    try {
        return JSON.parse(raw) as User;
    } catch (_) {
        return null;
    }
}

function saveUser(u: User): void {
    user = u;
    localStorage.setItem(USER_KEY, JSON.stringify(u));
}

function clearUser(): void {
    user = null;
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem(SELECTED_KEY);
}

function userToken(): string {
    return user ? user.token : "";
}

function userEmail(): string {
    return user ? user.email : "";
}

function userIsAuth(): boolean {
    return user !== null && user.token.length > 0;
}

function selectedContainer(): string {
    return localStorage.getItem(SELECTED_KEY) || "";
}

function selectContainer(id: string): void {
    localStorage.setItem(SELECTED_KEY, id);
}
