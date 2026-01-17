// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: "Toki",
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/pratyushtiwary/toki",
        },
      ],
      sidebar: [
        {
          label: "Tutorials",
          autogenerate: {
            directory: "tutorials",
          },
        },
        {
          label: "How-To Guides",
          autogenerate: {
            directory: "how-tos",
          },
        },
        {
          label: "Explanations",
          autogenerate: {
            directory: "explanations",
          },
        },
        {
          label: "References",
          link: "https://pkg.go.dev/github.com/pratyushtiwary/toki",
        },
      ],
    }),
  ],
});
