import { Component } from 'react';

export default class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error) {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div style={{ padding: 24, fontFamily: 'Segoe UI, sans-serif' }}>
          <h2>Algo falló al mostrar la pantalla</h2>
          <p style={{ color: '#b91c1c' }}>{String(this.state.error.message || this.state.error)}</p>
          <button type="button" onClick={() => { this.setState({ error: null }); window.location.reload(); }}>
            Recargar
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
