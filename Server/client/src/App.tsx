import * as React from 'react';
import { FaEnvelope } from 'react-icons/fa';
import './App.css';
import EmailLoginModal from './modal/EmailLoginModal';

declare abstract class cefSharp {}
export const isEmbedded = typeof cefSharp === 'object';


interface IntegrationInterface {
  requestStats(): string;
}

declare let core: IntegrationInterface;

interface State {
  modalVisible: boolean;
}

class App extends React.Component {
  state: State;

  constructor(props: {}) {
    super(props);
    this.state = {modalVisible: false};
  }

  onModalClose() {
    this.setState({modalVisible: false});
  }

  public render() {
    return (
      <div className="App">
        {!isEmbedded && <h1>You do not appear to be in a view embedded inside IA. Login will mostly fail (for now)</h1>}
        <h1 className="logo header">Sign-In for Online Backups for GD Item Assistant</h1>
        <i>Keeping your items safe.</i>

        <div className="login-container">
          <a className="btn btn-block btn-social btn-email" href="#"
             onClick={() => this.setState({modalVisible: true})}>
            <FaEnvelope /> Sign in with E-Mail
          </a>
        </div>
        {this.state.modalVisible ? <EmailLoginModal onClose={() => this.onModalClose()} visible={true}/> : ''}

        <div className="disclaimer">
          <b>By using this service, the following details will be stored about you:</b><br/>
          <ul>
            <li>Your e-mail address</li>
            <li>The data required to recreate your Grim Dawn items</li>
            <li>Your characters</li>
            <li>The date/time each item were uploaded</li>
            <li>Your IP address for ~24 hours (throttling excess traffic, prevent brute force logins)</li>
          </ul>
          Your email may in rare cases be utilized to contact you to help diagnose issues detected.
        </div>
        <br/><br/>

        <footer>
          <i><b>If you run into issues, hop unto the <a href="https://discord.com/invite/5wuCPbB"
                                                        rel="noopener noreferrer" target="_blank">IA
            discord.</a></b></i>
        </footer>
      </div>
    );
  }
}

export default App;
