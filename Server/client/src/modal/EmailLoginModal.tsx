import * as React from 'react';
import Modal from 'react-modal';
import { FaWindowClose } from 'react-icons/fa';
import './EmailLoginModal.css';
import EmailStage from './EmailStage';
import PincodeStage from './PincodeStage';

interface Props {
  visible: boolean;
  onClose: () => void;
}

enum Stage {
  Email, Code, InvalidCode
}

interface State {
  stage: Stage;
  email?: string;
  errorMessage?: string;
  code?: string;
  token?: string;
}

class EmailLoginModal extends React.Component<Props> {
  state: State;

  constructor(props: Props) {
    super(props);
    this.state = {stage: Stage.Email};
  }

  renderInvalidCode() {
    return (
      <div>
        <h2>You have entered an invalid verification code</h2>
        <br/>
        <br/>
        <br/>
        <button value="Close" className="btn btn-primary" onClick={() => this.props.onClose()}/>
      </div>
    );
  }

  onEmailStageComplete(email: string, token: string) {
    this.setState({stage: Stage.Code, email: email, token: token});
  }

  onCodeStageComplete(success: boolean, token?: string) {
    if (success) {
      // C# HttpUtility.ParseQueryString is utter shit and parses "?bug" as part of the URL.
      document.location.href = `https://token.iagd.evilsoft.net/?bug=1&email=${this.state.email}&token=${token}`;
    } else {
      this.setState({stage: Stage.InvalidCode});
    }
  }

  render() {
    let stage = this.state.stage;
    return (
      <div>
        <Modal
          isOpen={true}
          onRequestClose={() => this.props.onClose()}
          contentLabel="Verify your e-mail"
          ariaHideApp={false}
        >
          <div className="email-modal">
            <span className="modal-btn-close" onClick={() => this.props.onClose()}>
              <FaWindowClose />
            </span>

            {stage === Stage.Email && <EmailStage onCompletion={(email, token) => this.onEmailStageComplete(email, token)} />}
            {stage === Stage.Code && <PincodeStage
              onCompletion={(success: boolean, token?: string) => this.onCodeStageComplete(success, token)}
              email={this.state.email as string}
              token={this.state.token as string}
            />
            }
            {stage === Stage.InvalidCode && this.renderInvalidCode()}
          </div>
        </Modal>
      </div>
    );
  }
}

export default EmailLoginModal;

