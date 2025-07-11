import React, { useContext, useEffect, useState } from 'react';
import { Button, Divider, Form, Header, Image, Message, Modal } from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { API, copy, showError, showInfo, showNotice, showSuccess } from '../helpers';
import Turnstile from 'react-turnstile';
import { UserContext } from '../context/User';
import { onGitHubOAuthClicked, onLarkOAuthClicked } from './utils';

const PersonalSetting = () => {
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();

  const [inputs, setInputs] = useState({
    wechat_verification_code: '',
    email_verification_code: '',
    email: '',
    self_account_deletion_confirmation: ''
  });
  const [status, setStatus] = useState({});
  const [showWeChatBindModal, setShowWeChatBindModal] = useState(false);
  const [showEmailBindModal, setShowEmailBindModal] = useState(false);
  const [showAccountDeleteModal, setShowAccountDeleteModal] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const [affLink, setAffLink] = useState("");
  const [systemToken, setSystemToken] = useState("");

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);
      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
  }, []);

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval); // Clean up on unmount
  }, [disableButton, countdown]);

  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const generateAccessToken = async () => {
    const res = await API.get('/api/user/token');
    const { success, message, data } = res.data;
    if (success) {
      setSystemToken(data);
      setAffLink(""); 
      await copy(data);
      showSuccess(`Token has been reset and copied to clipboard`);
    } else {
      showError(message);
    }
  };

  const getAffLink = async () => {
    const res = await API.get('/api/user/aff');
    const { success, message, data } = res.data;
    if (success) {
      let link = `${window.location.origin}/register?aff=${data}`;
      setAffLink(link);
      setSystemToken("");
      await copy(link);
      showSuccess(`Invitation link has been copied to clipboard`);
    } else {
      showError(message);
    }
  };

  const handleAffLinkClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(`Invitation link has been copied to clipboard`);
  };

  const handleSystemTokenClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(`System token has been copied to clipboard`);
  };

  const deleteAccount = async () => {
    if (inputs.self_account_deletion_confirmation !== userState.user.username) {
      showError('Please enter your username to confirm deletion!');
      return;
    }

    const res = await API.delete('/api/user/self');
    const { success, message } = res.data;

    if (success) {
      showSuccess('Account has been deleted!');
      await API.get('/api/user/logout');
      userDispatch({ type: 'logout' });
      localStorage.removeItem('user');
      navigate('/login');
    } else {
      showError(message);
    }
  };

  const bindWeChat = async () => {
    if (inputs.wechat_verification_code === '') return;
    const res = await API.get(
      `/api/oauth/wechat/bind?code=${inputs.wechat_verification_code}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('WeChat account bound successfully!');
      setShowWeChatBindModal(false);
    } else {
      showError(message);
    }
  };

  const sendVerificationCode = async () => {
    setDisableButton(true);
    if (inputs.email === '') return;
    if (turnstileEnabled && turnstileToken === '') {
      showInfo('Please try again in a few seconds, Turnstile is checking the user environment!');
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('Verification code sent successfully, please check your email!');
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const bindEmail = async () => {
    if (inputs.email_verification_code === '') return;
    setLoading(true);
    const res = await API.get(
      `/api/oauth/email/bind?email=${inputs.email}&code=${inputs.email_verification_code}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess('Email account bound successfully!');
      setShowEmailBindModal(false);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  return (
    <div style={{ lineHeight: '40px' }}>
      <Header as='h3'>General Settings</Header>
      <Message>
        Note: The token generated here is for system management, not for requesting OpenAI-related services.
      </Message>
      <Button as={Link} to={`/user/edit/`}>
        Update Personal Information
      </Button>
      <Button onClick={generateAccessToken}>Generate System Access Token</Button>
      <Button onClick={getAffLink}>Copy Invitation Link</Button>
      <Button onClick={() => {
        setShowAccountDeleteModal(true);
      }}>Delete Personal Account</Button>
      
      {systemToken && (
        <Form.Input 
          fluid 
          readOnly 
          value={systemToken} 
          onClick={handleSystemTokenClick}
          style={{ marginTop: '10px' }}
        />
      )}
      {affLink && (
        <Form.Input 
          fluid 
          readOnly 
          value={affLink} 
          onClick={handleAffLinkClick}
          style={{ marginTop: '10px' }}
        />
      )}
      <Divider />
      <Header as='h3'>Account Binding</Header>
      {
        status.wechat_login && (
          <Button
            onClick={() => {
              setShowWeChatBindModal(true);
            }}
          >
            Bind WeChat Account
          </Button>
        )
      }
      <Modal
        onClose={() => setShowWeChatBindModal(false)}
        onOpen={() => setShowWeChatBindModal(true)}
        open={showWeChatBindModal}
        size={'mini'}
      >
        <Modal.Content>
          <Modal.Description>
            <Image src={status.wechat_qrcode} fluid />
            <div style={{ textAlign: 'center' }}>
              <p>
                Scan with WeChat to follow the official account, enter 'verification code' to get the code (valid for 3 minutes)
              </p>
            </div>
            <Form size='large'>
              <Form.Input
                fluid
                placeholder='Verification Code'
                name='wechat_verification_code'
                value={inputs.wechat_verification_code}
                onChange={handleInputChange}
              />
              <Button color='' fluid size='large' onClick={bindWeChat}>
                Bind
              </Button>
            </Form>
          </Modal.Description>
        </Modal.Content>
      </Modal>
      {
        status.github_oauth && (
          <Button onClick={()=>{onGitHubOAuthClicked(status.github_client_id)}}>Bind GitHub Account</Button>
        )
      }
      {
        status.lark_client_id && (
          <Button onClick={()=>{onLarkOAuthClicked(status.lark_client_id)}}>Bind Lark Account</Button>
        )
      }
      <Button
        onClick={() => {
          setShowEmailBindModal(true);
        }}
      >
        Bind Email Address
      </Button>
      <Modal
        onClose={() => setShowEmailBindModal(false)}
        onOpen={() => setShowEmailBindModal(true)}
        open={showEmailBindModal}
        size={'tiny'}
        style={{ maxWidth: '450px' }}
      >
        <Modal.Header>Bind Email Address</Modal.Header>
        <Modal.Content>
          <Modal.Description>
            <Form size='large'>
              <Form.Input
                fluid
                placeholder='Enter email address'
                onChange={handleInputChange}
                name='email'
                type='email'
                action={
                  <Button onClick={sendVerificationCode} disabled={disableButton || loading}>
                    {disableButton ? `Resend (${countdown})` : 'Get Verification Code'}
                  </Button>
                }
              />
              <Form.Input
                fluid
                placeholder='Verification Code'
                name='email_verification_code'
                value={inputs.email_verification_code}
                onChange={handleInputChange}
              />
              {turnstileEnabled ? (
                <Turnstile
                  sitekey={turnstileSiteKey}
                  onVerify={(token) => {
                    setTurnstileToken(token);
                  }}
                />
              ) : (
                <></>
              )}
              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem' }}>
              <Button
                color=''
                fluid
                size='large'
                onClick={bindEmail}
                loading={loading}
              >
                Confirm Binding
              </Button>
              <div style={{ width: '1rem' }}></div> 
              <Button
                fluid
                size='large'
                onClick={() => setShowEmailBindModal(false)}
              >
                Cancel
              </Button>
              </div>
            </Form>
          </Modal.Description>
        </Modal.Content>
      </Modal>
      <Modal
        onClose={() => setShowAccountDeleteModal(false)}
        onOpen={() => setShowAccountDeleteModal(true)}
        open={showAccountDeleteModal}
        size={'tiny'}
        style={{ maxWidth: '450px' }}
      >
        <Modal.Header>Dangerous Operation</Modal.Header>
        <Modal.Content>
        <Message>You are deleting your own account, which will clear all data and cannot be recovered</Message>
          <Modal.Description>
            <Form size='large'>
              <Form.Input
                fluid
                placeholder={`Enter your username ${userState?.user?.username} to confirm deletion`}
                name='self_account_deletion_confirmation'
                value={inputs.self_account_deletion_confirmation}
                onChange={handleInputChange}
              />
              {turnstileEnabled ? (
                <Turnstile
                  sitekey={turnstileSiteKey}
                  onVerify={(token) => {
                    setTurnstileToken(token);
                  }}
                />
              ) : (
                <></>
              )}
              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem' }}>
                <Button
                  color='red'
                  fluid
                  size='large'
                  onClick={deleteAccount}
                  loading={loading}
                >
                  Confirm Delete
                </Button>
                <div style={{ width: '1rem' }}></div>
                <Button
                  fluid
                  size='large'
                  onClick={() => setShowAccountDeleteModal(false)}
                >
                  Cancel
                </Button>
              </div>
            </Form>
          </Modal.Description>
        </Modal.Content>
      </Modal>
    </div>
  );
};

export default PersonalSetting;
