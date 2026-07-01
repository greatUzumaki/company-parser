export function Footer() {
  return (
    <footer className="mt-10 border-t border-slate-200 py-6 text-center text-xs text-slate-500 dark:border-slate-800 dark:text-slate-400">
      <p>
        Data ©{" "}
        <a
          href="https://www.openstreetmap.org/copyright"
          target="_blank"
          rel="noopener noreferrer"
          className="font-medium text-emerald-600 hover:underline dark:text-emerald-400"
        >
          OpenStreetMap contributors
        </a>{" "}
        (ODbL)
      </p>
    </footer>
  );
}
