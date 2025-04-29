// app/utils/authStorage.ts
import * as SecureStore from 'expo-secure-store';

// Keys for storing different pieces of auth data
const KEYS = {
  ACCESS_TOKEN: 'navik_access_token',
  REFRESH_TOKEN: 'navik_refresh_token',
  USER_ID: 'navik_user_id',
  USER_TYPE: 'navik_user_type',
  USER_DATA: 'navik_user_data',
  EXPIRES_AT: 'navik_token_expires_at'
};

// Auth response interface matching your API
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user_id: string;
  user_type: string;
}

// Store auth data received from login/register
export async function storeAuthData(authResponse: AuthResponse): Promise<void> {
  try {
    const expiresAt = Date.now() + authResponse.expires_in * 1000;
    
    await SecureStore.setItemAsync(KEYS.ACCESS_TOKEN, authResponse.access_token);
    await SecureStore.setItemAsync(KEYS.REFRESH_TOKEN, authResponse.refresh_token);
    await SecureStore.setItemAsync(KEYS.USER_ID, authResponse.user_id);
    await SecureStore.setItemAsync(KEYS.USER_TYPE, authResponse.user_type);
    await SecureStore.setItemAsync(KEYS.EXPIRES_AT, expiresAt.toString());
  } catch (error) {
    console.error('Error storing auth data:', error);
    throw new Error('Failed to store authentication data securely');
  }
}

// Store user profile data 
export async function storeUserData(userData: any): Promise<void> {
  try {
    // SecureStore can only store strings, so we need to serialize the object
    const userDataString = JSON.stringify(userData);
    await SecureStore.setItemAsync(KEYS.USER_DATA, userDataString);
  } catch (error) {
    console.error('Error storing user data:', error);
    throw new Error('Failed to store user data securely');
  }
}

// Get the current access token
export async function getAccessToken(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.ACCESS_TOKEN);
}

// Get the refresh token
export async function getRefreshToken(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.REFRESH_TOKEN);
}

// Get user ID
export async function getUserId(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.USER_ID);
}

// Get user type (driver or customer)
export async function getUserType(): Promise<string | null> {
  return SecureStore.getItemAsync(KEYS.USER_TYPE);
}

// Get user profile data
export async function getUserData(): Promise<any | null> {
  try {
    const userDataString = await SecureStore.getItemAsync(KEYS.USER_DATA);
    if (!userDataString) return null;
    return JSON.parse(userDataString);
  } catch (error) {
    console.error('Error retrieving user data:', error);
    return null;
  }
}

// Check if the token is expired
export async function isTokenExpired(): Promise<boolean> {
  try {
    const expiresAtStr = await SecureStore.getItemAsync(KEYS.EXPIRES_AT);
    if (!expiresAtStr) return true;
    
    const expiresAt = parseInt(expiresAtStr, 10);
    return Date.now() >= expiresAt;
  } catch (error) {
    console.error('Error checking token expiration:', error);
    return true; // Assume expired on error
  }
}

// Clear all auth data on logout
export async function clearAuthData(): Promise<void> {
  try {
    await SecureStore.deleteItemAsync(KEYS.ACCESS_TOKEN);
    await SecureStore.deleteItemAsync(KEYS.REFRESH_TOKEN);
    await SecureStore.deleteItemAsync(KEYS.USER_ID);
    await SecureStore.deleteItemAsync(KEYS.USER_TYPE);
    await SecureStore.deleteItemAsync(KEYS.USER_DATA);
    await SecureStore.deleteItemAsync(KEYS.EXPIRES_AT);
  } catch (error) {
    console.error('Error clearing auth data:', error);
  }
}

// Check if user is authenticated
export async function isAuthenticated(): Promise<boolean> {
  try {
    const token = await getAccessToken();
    const expired = await isTokenExpired();
    return !!token && !expired;
  } catch (error) {
    return false;
  }
}
