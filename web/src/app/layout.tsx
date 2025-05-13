import type { Viewport } from "next";
import React from "react";

import "@/app/styles/globals.css";
import { cn } from "@/lib/utils";

import { fontMono, fontSans, fontSerif } from "@/lib/fonts";
import { ThemeProvider } from "@/components/providers";
// import TailwindIndicator from "@/components/tailwind-indicator";
import { Toaster } from "@/components/ui/toaster";
import "./styles/Treant.css";

import { env } from "@/env";

export const viewport: Viewport = {
  colorScheme: "dark light",
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "white" },
    { media: "(prefers-color-scheme: dark)", color: "black" },
  ],
};

function Analytics() {
  if (env.NEXT_ANALYTICS_UMAMI_ID) {
    return (
      <script
        defer
        src="https://analytics.us.umami.is/script.js"
        data-website-id={env.NEXT_ANALYTICS_UMAMI_ID}
      ></script>
    );
  }
  return null;
}

export default function RootLayout({
  children,
  params: { locale },
}: {
  children: React.ReactNode;
  params: { locale: string };
}) {
  return (
    <html lang={locale} suppressHydrationWarning>
      <head>
        <Analytics />
        <script src="/treant/vendor/raphael.js" defer />
        <script src="/treant/Treant.min.js" defer />
        {/* <link
          rel="stylesheet"
          href="https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;700&family=Fira+Mono:wght@400;700&family=Noto+Serif+SC:wght@400;700&display=swap"
        /> */}
      </head>
      <body
        className={cn(
          "min-h-screen bg-background font-sans antialiased",
          fontSans.variable,
          fontMono.variable,
          fontSerif.variable
        )}
      >
        {/* <TailwindIndicator /> */}
        <ThemeProvider
          attribute="class"
          defaultTheme="light"
          enableSystem
          disableTransitionOnChange
        >
          <main>{children}</main>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
