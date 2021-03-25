import * as React from 'react';
import './EmailLoginModal.css';
import '../spinner.css';

interface Props {
  onCompletion: (email: string, token: string) => void;
}

interface State {
  email?: string;
  errorMessage?: string;
  loading: boolean;
}

class EmailStage extends React.Component<Props> {
  state: State;

  constructor(props: Props) {
    super(props);
    this.state = { loading: false };
  }

  onSendEmail() {
    const email = this.state.email as string;

    if (this.validateEmail(email)) {
      let self = this;
      const uri = 'https://api.iagd.evilsoft.net/login';
      this.setState({loading: true});

      fetch(`${uri}?email=${email}`, {
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          }
        }
      )
        .then((response) => {
          if (!response.ok) {
            console.log(response);
            throw Error(`Got response ${response.status}, ${response.statusText}`);
          }
          return response;
        })
        .then((response) => response.json())
        .then((json) => {
          if (json.key !== undefined) {
            this.props.onCompletion(email, json.key);
          }
          else {
            console.warn('Attempted to fetch token for email, but token was undefined.');
          }
          this.setState({loading: false});
        })
        .catch((error) => {
          console.warn(error);
          self.setState({loading: false, errorMessage: `${error}`});
        });

    } else {
      this.setState({errorMessage: 'This is not a valid e-mail address'});
    }
  }

  onEmailChange(email: string) {
    if (this.state.errorMessage) {
      this.setState({errorMessage: undefined});
    }
    else {
      this.setState({email: email});
    }
  }

  validateEmail(email: string) {
    var re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
    return re.test(String(email).toLowerCase());
  }

  render() {
    return (
      <div>
        <h2>Please enter your e-mail address</h2>

        <span> A Pin Code will be sent to verify your identity</span>
        <div className="email-form">
          <div className="form-group">
            <input
              className="form-control"
              autoFocus={true}
              type="email"
              placeholder="Your e-mail address"
              required={true}
              max="255"
              onChange={(e) => this.onEmailChange(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' ? this.onSendEmail() : ''}
            />
            {this.state.errorMessage && <div className="alert alert-warning">{this.state.errorMessage}</div>}
            {!this.state.loading && <button className="form-control btn btn-primary" onClick={() => this.onSendEmail()}>Send</button>}

            {this.state.loading && <div className="loader-container">
              <div className="loader"></div>
            </div>}
          </div>
        </div>
      </div>
    );
  }
}

export default EmailStage;

