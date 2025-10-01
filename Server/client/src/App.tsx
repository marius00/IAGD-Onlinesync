import * as React from 'react';
import { FaEnvelope } from 'react-icons/fa';
import './App.css';
import EmailLoginModal from './modal/EmailLoginModal';


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
        <h1 className="logo header">Sign-In for Online Backups for GD Item Assistant</h1>
        <i>Keeping your items safe.</i>
        <hr/>


        {this.state.modalVisible ? <EmailLoginModal visible={true}/> : ''}

        {!this.state.modalVisible ? <div>
          <div className="login-container">
            <a className="btn btn-block btn-social btn-email" href="#"
               onClick={() => this.setState({modalVisible: true})}>
              <FaEnvelope/> Sign in with E-Mail
            </a>
          </div>
        <div className="disclaimer">
          <b>By using this service, the following details will be stored about you:</b><br/>
          <ul>
            <li>Your e-mail address</li>
            <li>The data required to recreate your Grim Dawn items</li>
            <li>Your characters</li>
            <li>Your transfer stash file</li>
            <li>The date/time each item were uploaded</li>
            <li>Your IP address for ~24 hours (throttling excess traffic, prevent brute force logins)</li>
          </ul>
        </div>
        </div> : ''}
        <br/><br/>

        <footer>
          <i><b>If you run into issues, hop unto the <a href="https://discord.com/invite/5wuCPbB" rel="noopener noreferrer" target="_blank">IA discord.</a></b></i>
        </footer>
      </div>
    );
  }
}

export default App;
