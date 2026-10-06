import { config } from "@remotion/eslint-config-flat";

// The Remotion rules police compositions. hub/src/components/ui/ is shadcn's
// code for a web page: its CSS transitions never render to video.
export default [
  ...(Array.isArray(config) ? config : [config]),
  {
    files: ["hub/src/components/ui/**/*.tsx"],
    rules: {
      "@remotion/non-pure-animation": "off",
      "@remotion/warn-native-media-tag": "off",
    },
  },
];
