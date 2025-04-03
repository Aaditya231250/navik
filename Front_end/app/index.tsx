// import { Text } from "react-native"; // Importing components from React Native
// import { SafeAreaView } from "react-native-safe-area-context";
// import { Redirect } from "expo-router";
// import { useAuth } from "@clerk/clerk-expo";
//
// const Home = () => {
//   // Functional component
//   const { isSignedIn } = useAuth();
//
//   if (isSignedIn) {
//     return <Redirect href={"/"} />;
//   }
//
//   return <Redirect href={"/(auth)/welcome"} />;
// };
//
// export default Home; // Exporting the component

import { useAuth } from "@clerk/clerk-expo";
import { Redirect } from "expo-router";

const Page = () => {
  const { isSignedIn } = useAuth();

  if (isSignedIn) return <Redirect href="/(root)/(tabs)/home" />;

  return <Redirect href="/(auth)/welcome" />;
};

export default Page;
