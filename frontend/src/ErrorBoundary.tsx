import React from "react";

type Props = {
  children: React.ReactNode;
};

type State = {
  hasError: boolean;
  message: string;
};

export class ErrorBoundary extends React.Component<Props, State> {
  state: State = { hasError: false, message: "" };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error.message || String(error) };
  }

  componentDidCatch(error: Error): void {
    console.error("UI crashed:", error);
  }

  render(): React.ReactNode {
    if (this.state.hasError) {
      return (
        <div style={{ padding: 24, color: "#fff" }}>
          <h1>UI Error</h1>
          <p>The page crashed while rendering.</p>
          <pre style={{ whiteSpace: "pre-wrap" }}>{this.state.message}</pre>
        </div>
      );
    }
    return this.props.children;
  }
}
