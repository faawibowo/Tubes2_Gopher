import React from "react";
import { SiteHeader } from "@/components/layouts/site-header";
import { Toaster } from "@/components/ui/toaster";

export default async function LobyLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="relative flex min-h-screen flex-col leading-relaxed">
      <SiteHeader />
      <main>
        <div className="relative flex min-h-[80vh] flex-col">
          {children}
          <Toaster></Toaster>
        </div>
      </main>
    </div>
  );
}
