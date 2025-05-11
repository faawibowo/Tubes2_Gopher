// import type { MainNavItem } from "@/types";

import type { Icons } from "@/components/icons";

export interface NavItem {
  id: string;
  title?: string;
  href?: string;
  disabled?: boolean;
  external?: boolean;
  icon?: keyof typeof Icons;
  label?: string;
  description?: string;
  active?: boolean;
}

export interface NavItemWithChildren extends NavItem {
  items: NavItemWithChildren[];
}
export interface NavItemWithOptionalChildren extends NavItem {
  items?: NavItemWithChildren[];
}

export type MainNavItem = NavItemWithOptionalChildren;
export type SidebarNavItem = NavItemWithChildren;

export type SiteConfig = typeof siteConfig;
export const siteConfig = {
  mainNav: [
    {
      id: "Kocak",
      title: "Articles",
      items: [
        {
          id: "ArticleList",
          href: "",
          title: "Farrel ganteng",
          description: "omg 😎",
          items: [],
        },
        {
          id: "ArticleTags",
          href: "",
          title: "Brian Tamin ganteng",
          description: "aku sayang kamu deh 😍",
          items: [],
        },
      ],
    },
    {
      id: "About",
      title: "About",
      href: "/page/about",
    },
  ] satisfies MainNavItem[],
};
