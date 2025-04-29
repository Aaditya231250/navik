import { Slot } from "expo-router";
import { SafeAreaView, StyleSheet } from "react-native";

export default function AuthLayout() {
  return (
    <SafeAreaView style={styles.container}>
      <Slot />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
  },
});
