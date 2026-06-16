declare module '*.css';
declare module '*.svg' {
  const content: string;
  export default content;
}

interface Window {
  runtime: any;
  go: any;
}
