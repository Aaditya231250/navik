// app/utils/api.ts
import { getAccessToken, getRefreshToken, storeAuthData, clearAuthData, storeUserData } from './authStorage';

const API_BASE_URL = "http://172.31.115.2/api"; // Replace with your API URL

// Custom fetch function that automatically adds auth headers
export async function authFetch(
  endpoint: string, 
  options: RequestInit = {}
): Promise<Response> {
  const token = await getAccessToken();
  
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    ...options.headers,
  };
  
  const config = {
    ...options,
    headers,
  };
  
  // Make the request
  const response = await fetch(`${API_BASE_URL}${endpoint}`, config);
  
  // Handle 401 Unauthorized errors (expired token)
  if (response.status === 401) {
    const refreshed = await refreshAccessToken();
    
    // If refresh successful, retry the original request
    if (refreshed) {
      const newToken = await getAccessToken();
      const newHeaders = {
        ...headers,
        'Authorization': `Bearer ${newToken}`,
      };
      
      return fetch(`${API_BASE_URL}${endpoint}`, {
        ...config,
        headers: newHeaders,
      });
    }
    
    // If refresh failed, force logout
    await clearAuthData();
    // You might want to redirect to login screen here
  }
  
  return response;
}

// Function to refresh the access token
async function refreshAccessToken(): Promise<boolean> {
  try {
    const refreshToken = await getRefreshToken();
    if (!refreshToken) return false;
    
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    
    if (response.ok) {
      const data = await response.json();
      await storeAuthData(data);
      return true;
    }
    
    return false;
  } catch (error) {
    console.error('Error refreshing token:', error);
    return false;
  }
}

// Authentication API calls
export const authAPI = {
  login: async (email: string, password: string) => {
    const response = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    });
    
    const data = await response.json();
    if (response.ok) {
      await storeAuthData(data);
    }

    // After successful login, fetch and store profile data
    try {
      const userType = data.user_type;
      const profileEndpoint = userType === 'driver' ? '/driver/profile' : '/customer/profile';
      
      const profileResponse = await fetch(`${API_BASE_URL}${profileEndpoint}`, {
        headers: {
          'Authorization': `Bearer ${data.access_token}`,
          'Content-Type': 'application/json',
        },
      });
      
      if (profileResponse.ok) {
        const profileData = await profileResponse.json();
        await storeUserData(profileData);
      }
    } catch (error) {
      console.error('Failed to fetch profile data:', error);
    }
    return { success: response.ok, data };
  },
  
  register: async (formData: any) => {
    const response = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(formData),
    });
    
    return { success: response.ok, data: await response.json() };
  },
  
  logout: async () => {
    // Optionally call an API endpoint to invalidate the token on the server
    try {
      await authFetch('/auth/logout', { method: 'POST' });
    } catch (error) {
      console.error('Error logging out:', error);
    } finally {
      await clearAuthData();
    }
  }
};

// Example domain-specific API calls that use authFetch
export const userAPI = {
  getProfile: async () => {
    const response = await authFetch('/user/profile');
    if (response.ok) {
      const userData = await response.json();
      // Store user data in SecureStore
      await storeUserData(userData);
      return userData;
    }
    throw new Error('Failed to get user profile');
  },
  
  updateProfile: async (profileData: any) => {
    const response = await authFetch('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(profileData),
    });
    return response.ok;
  }
};

export const rideAPI = {
  requestRide: async (rideData: any) => {
    const response = await authFetch('/rides', {
      method: 'POST',
      body: JSON.stringify(rideData),
    });
    return response.ok ? await response.json() : null;
  },
  
  getRideHistory: async () => {
    const response = await authFetch('/rides/history');
    return response.ok ? await response.json() : [];
  }
};
