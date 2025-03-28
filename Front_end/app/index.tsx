import { Text } from "react-native"; // Importing components from React Native
import { SafeAreaView } from "react-native-safe-area-context";
import { Redirect } from "expo-router";
const Home = () => {
  // Functional component
  return <Redirect href={"/(auth)/welcome"} />;
};

export default Home; // Exporting the component
