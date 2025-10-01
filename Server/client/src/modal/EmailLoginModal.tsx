import * as React from 'react';
import './EmailLoginModal.css';
import EmailStage from './EmailStage';
import PincodeStage from './PincodeStage';

interface Props {
  visible: boolean;
}

enum Stage {
  Email, Code, InvalidCode, PendingIntegration, CompletedIntegration, Error, Expired
}

interface State {
  stage: Stage;
  email?: string;
  errorMessage?: string;
  code?: string;
  token?: string;
  dots: number;
  numPollsDone: number;
}

class EmailLoginModal extends React.Component<Props> {
  state: State;

  constructor(props: Props) {
    super(props);
    this.state = {stage: Stage.Email, dots: 1, numPollsDone: 0};
  }

  renderWaitingForIntegration() {
    return (
      <div>
        <h2>Code verified</h2>
        <p>
          Waiting for Item Assistant to connect to the cloud{".".repeat(this.state.dots)}
        </p>
      </div>
    );
  }

  renderInvalidCode() {
    return (
      <div>
        <h2>You have entered an invalid verification code</h2>
        <br/>
      </div>
    );
  }

  onEmailStageComplete(email: string, token: string) {
    this.setState({stage: Stage.Code, email: email, token: token});
  }

  onCodeStageComplete(success: boolean, token?: string) {
    if (success) {
      if (new URLSearchParams(document.location.search).get('token')) {
        // Newer version of IAGD uses Edge, and we'll poll for the result.
        this.setState({stage: Stage.PendingIntegration});
        this.onIntegrationPullComplete(token);
      } else {
        // CefSharp redirect hook picks this up
        // C# HttpUtility.ParseQueryString is utter shit and parses "?bug" as part of the URL.
        document.location.href = `https://token.iagd.evilsoft.net/?bug=1&email=${this.state.email}&token=${token}`;
      }
    } else {
      this.setState({stage: Stage.InvalidCode});
    }
  }

  onIntegrationPullComplete(token?: string) {
    if (this.state.numPollsDone > 60*15) {
      this.setState({stage: Stage.Expired});
      return;
    } else if (this.state.stage !== Stage.PendingIntegration) {
      console.log("Status is not PendingIntegration, aborting poll");
      return;
    }

    fetch('https://api.iagd.evilsoft.net/status', {
        method: 'POST',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: `token=${token}`
      }
    )
      .then((response) => {
        if (!response.ok) {
          console.log(response);
          this.setState({state: Stage.Error});
          throw Error(`Got response ${response.status}, ${response.statusText}`);
        }
        return response;
      })
      .then((response) => response.json())
      .then((json) => {
        if (json.status !== undefined) {
          if (json.status === 'COMPLETED') {
            this.setState({stage: Stage.CompletedIntegration});
          } else {
            console.log("Status is", json.status, ", waiting 1 second.")
            this.setState({dots: (this.state.dots + 1) % 4, numPollsDone: this.state.numPollsDone + 1});
            setTimeout(() => this.onIntegrationPullComplete(token), 1000);
          }
        }
        else {
          console.warn('The result status was undefined.');
          this.setState({state: Stage.Error});
        }
      })
      .catch((error) => {
        console.warn(error);
        this.setState({state: Stage.Error});
      });
  }

  render() {
    let stage = this.state.stage;
    return (
      <div>
        <div>
          <div className="email-modal">
            {stage === Stage.Email && <EmailStage onCompletion={(email, token) => this.onEmailStageComplete(email, token)} />}
            {stage === Stage.Code && <PincodeStage
              onCompletion={(success: boolean, token?: string) => this.onCodeStageComplete(success, token)}
              email={this.state.email as string}
              token={this.state.token as string}
            />
            }
            {stage === Stage.InvalidCode && this.renderInvalidCode()}
            {stage === Stage.PendingIntegration && this.renderWaitingForIntegration()}
            {stage === Stage.Error && <div>
                <h1>Something went wrong</h1>
                <br/>
                Something went wrong.. Ask for help on the IA discord or try again later.
            </div>}
            {stage === Stage.Expired && <div>
                <h1>Something went wrong</h1>
                <br/>
                Something went wrong, Item Assistant did not connect to the cloud
            </div>}
            {stage === Stage.CompletedIntegration
              && <div>
                    <h1>Login successful</h1>
                    <br/>
                    You can safely close this window.
            </div>
            }
          </div>
        </div>
      </div>
    );
  }
}

export default EmailLoginModal;

