import type { Metadata } from "next";
import "./globals.css";
import Shell from "@/components/Shell";
import ThemeProvider from "@/components/ThemeProvider";

export const metadata: Metadata = {
  title: "Google Keep Clone",
  description: "A Google Keep clone built with Next.js",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased" suppressHydrationWarning>
      <body className="min-h-full dark:bg-[#202124]">
        <ThemeProvider>
          <Shell>{children}</Shell>
        </ThemeProvider>
      </body>
    </html>
  );
}
