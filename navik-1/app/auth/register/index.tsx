import React, { useState } from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Alert,
} from "react-native";
import { useRouter } from "expo-router";
import { authAPI } from "@/utils/api";

export default function RegisterScreen() {
  const router = useRouter();
  const [form, setForm] = useState({
    first_name: "",
    last_name: "",
    email: "",
    password: "",
    phone: "",
    user_type: "customer",
  });
  const [loading, setLoading] = useState(false);

  const handleChange = (field, value) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleRegister = async () => {
    const { first_name, last_name, email, password, phone, user_type } = form;
    
    if (!first_name || !last_name || !email || !password || !phone) {
      Alert.alert("Error", "Please fill all fields");
      return;
    }
    
    setLoading(true);
    try {
      const { success, data } = await authAPI.register({
        first_name,
        last_name,
        email,
        password,
        phone,
        user_type
      });
      
      if (success) {
        Alert.alert("Success", "Registration successful", [
          { text: "OK", onPress: () => router.push("/auth/login") },
        ]);
      } else {
        Alert.alert("Error", data.message || "Registration failed");
      }
    } catch (error) {
      Alert.alert("Error", "Network error. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : undefined}
        style={{ flex: 1 }}
      >
        <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 20 }}>
          <Text style={styles.title}>NAVIK</Text>
          <Text style={styles.subtitle}>Create Account</Text>

          <TextInput
            style={styles.input}
            placeholder="First Name"
            value={form.first_name}
            onChangeText={(text) => handleChange("first_name", text)}
          />
          <TextInput
            style={styles.input}
            placeholder="Last Name"
            value={form.last_name}
            onChangeText={(text) => handleChange("last_name", text)}
          />
          <TextInput
            style={styles.input}
            placeholder="Email Address"
            keyboardType="email-address"
            autoCapitalize="none"
            value={form.email}
            onChangeText={(text) => handleChange("email", text)}
          />
          <TextInput
            style={styles.input}
            placeholder="Password"
            secureTextEntry
            value={form.password}
            onChangeText={(text) => handleChange("password", text)}
          />
          <TextInput
            style={styles.input}
            placeholder="Phone Number (+countrycode)"
            keyboardType="phone-pad"
            value={form.phone}
            onChangeText={(text) => handleChange("phone", text)}
          />

          {/* User type toggle */}
          <View style={{ flexDirection: "row", justifyContent: "space-around", marginVertical: 16 }}>
            <TouchableOpacity
              style={[
                styles.userTypeBtn,
                form.user_type === "customer" && styles.userTypeSelected,
              ]}
              onPress={() => handleChange("user_type", "customer")}
            >
              <Text
                style={[
                  styles.userTypeText,
                  form.user_type === "customer" && styles.userTypeTextSelected,
                ]}
              >
                Customer
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[
                styles.userTypeBtn,
                form.user_type === "driver" && styles.userTypeSelected,
              ]}
              onPress={() => handleChange("user_type", "driver")}
            >
              <Text
                style={[
                  styles.userTypeText,
                  form.user_type === "driver" && styles.userTypeTextSelected,
                ]}
              >
                Driver
              </Text>
            </TouchableOpacity>
          </View>

          <TouchableOpacity
            style={[styles.button, loading && { opacity: 0.7 }]}
            onPress={handleRegister}
            disabled={loading}
          >
            <Text style={styles.buttonText}>{loading ? "Creating Account..." : "SIGN UP"}</Text>
          </TouchableOpacity>

          <TouchableOpacity onPress={() => router.push("/auth/login")}>
            <Text style={styles.linkText}>Already have an account? Sign In</Text>
          </TouchableOpacity>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#fff" },
  title: { fontSize: 36, fontWeight: "bold", textAlign: "center", marginBottom: 8, marginTop: 20 },
  subtitle: { fontSize: 24, textAlign: "center", marginBottom: 24 },
  input: {
    height: 50,
    borderColor: "#ddd",
    borderWidth: 1,
    borderRadius: 8,
    marginBottom: 16,
    paddingHorizontal: 16,
    fontSize: 16,
  },
  button: {
    backgroundColor: "#00a5a0",
    height: 50,
    borderRadius: 8,
    justifyContent: "center",
    alignItems: "center",
    marginBottom: 12,
  },
  buttonText: { color: "#fff", fontSize: 18, fontWeight: "600" },
  linkText: { textAlign: "center", color: "#00a5a0", marginTop: 12, fontSize: 16 },
  userTypeBtn: {
    flex: 1,
    borderWidth: 1,
    borderColor: "#ddd",
    borderRadius: 8,
    paddingVertical: 12,
    marginHorizontal: 5,
    alignItems: "center",
  },
  userTypeSelected: { backgroundColor: "#00a5a0", borderColor: "#00a5a0" },
  userTypeText: { fontSize: 16, color: "#333" },
  userTypeTextSelected: { color: "#fff", fontWeight: "600" },
});
