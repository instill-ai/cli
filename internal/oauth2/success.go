package oauth2

const oauthSuccessPage = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Instill - Login Success</title>
    <meta name="description" content="Instill CLI login success page" />
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link
      rel="preconnect"
      href="https://fonts.googleapis.com"
      as="font"
      crossOrigin="anonymous"
    />
    <link
      rel="preconnect"
      href="https://fonts.gstatic.com"
      as="font"
      crossOrigin="anonymous"
    />
    <link
      href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&display=swap"
      rel="stylesheet"
    />
    <style>
      body {
        margin: 0px;
      }
      .container {
        display: flex;
      }
      .main {
        display: grid;
        min-height: 100vh;
        width: 100%;
        grid-template-columns: repeat(1, minmax(0, 1fr));
      }
      .text-container {
        display: flex;
      }
      .text-sub-container {
        display: flex;
        flex-direction: column;
        margin: auto;
        max-width: 375px;
      }
      .title {
        margin-top: 0px;
        margin-bottom: 40px;
        font-size: 36px;
        font-family: "IBM Plex Sans", sans-serif;
        text-align: center;
      }
      .text {
        font-family: "IBM Plex Sans", sans-serif;
        font-size: 16px;
        line-height: 24px;
        text-align: center;
        margin: 0;
      }
      .instill-logo-black {
        margin: 0 auto;
        width: 240px;
        margin-bottom: 40px;
      }
      .main-logo-container {
        display: none;
        background-color: #1a1a1a;
      }
      .instill-logo-white {
        margin: auto;
        width: 400px;
      }
      @media (min-width: 1280px) {
        .main {
          grid-template-columns: repeat(2, minmax(0, 1fr))
        }
        .main-logo-container{
          display: flex;
        }
        .instill-logo-black {
          display: none;
        }
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div class="main">
        <div class="main-logo-container">
          <svg class="instill-logo-white" viewBox="0 0 861 275" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M203.289 59.7815H59.7906V215.184H215.273V59.7815H203.289ZM71.7529 71.735H83.7082V107.602H71.7529V71.735ZM71.7529 119.563H83.7082V155.299H71.7529V119.563ZM107.626 203.258H71.7529V167.252H83.7082V167.391H95.6636V167.252H107.626V203.258ZM107.626 155.299H95.6636V119.563H107.626V155.299ZM107.626 107.602H95.6636V71.735H107.543V83.7995H107.626V107.602ZM119.581 119.563H131.543V155.299H119.581V119.563ZM119.581 167.252H131.543V179.046H119.581V167.252ZM155.322 191.332H131.682V179.275H143.478V167.252H155.274L155.322 191.332ZM155.322 155.333H143.527V119.563H155.322V155.333ZM155.322 95.6004H143.527V107.637H119.581V83.7995H131.474V95.5934H143.429V83.5706H131.543V71.735H155.322V95.6004ZM179.379 203.293H167.278V167.287H179.379V203.293ZM203.289 203.293H191.334V167.287H203.289V203.293ZM203.289 155.333H167.278V119.563H203.289V155.333ZM203.289 107.637H167.278V95.6489H203.289V107.637ZM203.289 83.7301H167.278V71.735H203.289V83.7301Z" fill="white"/>
            <path d="M705.539 155.437H693.584V167.391H705.539V155.437Z" fill="white"/>
            <path d="M753.048 107.644V107.609H729.457V119.563H741.419V143.477H717.501V131.523H729.457V119.563H717.501V131.523H705.539V155.437H710.84H717.501H741.419V167.391H753.374V107.644H753.048Z" fill="white"/>
            <path d="M801.209 119.563V107.609H765.336V119.563H777.292V155.437H765.336V167.391H801.209V155.437H789.247V119.563H801.209Z" fill="white"/>
            <path d="M406.586 131.523V119.563H442.459V107.609H406.586H394.624V143.477H406.586H442.459V155.437H454.414V131.523H442.459H406.586Z" fill="white"/>
            <path d="M442.16 155.437H394.624V167.391H442.16V155.437Z" fill="white"/>
            <path d="M275.042 119.563H286.998V155.437H275.042V167.391H310.915V155.437H298.96V119.563H310.915V107.609H275.042V119.563Z" fill="white"/>
            <path d="M454.414 119.563H478.332V167.391H490.294V119.563H514.489V107.609H454.414V119.563Z" fill="white"/>
            <path d="M526.306 119.563H538.192V155.437H526.306V167.391H562.04V155.437H550.154V119.563H562.04V107.609H526.306V119.563Z" fill="white"/>
            <path d="M585.958 107.609H574.002V155.437V167.099V167.391H609.875V155.437H585.958V107.609Z" fill="white"/>
            <path d="M633.793 107.609H621.831V155.437V167.099V167.391H657.711V155.437H633.793V107.609Z" fill="white"/>
            <path d="M334.833 119.563V107.609H322.878V119.563V131.523V167.391H334.833V131.523H346.795V119.563H334.833Z" fill="white"/>
            <path d="M358.744 131.523H346.788V143.477H358.744V131.523Z" fill="white"/>
            <path d="M370.706 143.477H358.751V155.437H370.706V167.391H382.668V155.437V143.477V107.609H370.706V143.477Z" fill="white"/>
            </svg>
        </div>
        <div class="text-container">
          <div class="text-sub-container">
            <svg class="instill-logo-black" viewBox="0 0 861 275" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M203.289 59.7814H59.7906V215.184H215.273V59.7814H203.289ZM71.7529 71.7349H83.7082V107.602H71.7529V71.7349ZM71.7529 119.563H83.7082V155.298H71.7529V119.563ZM107.626 203.258H71.7529V167.252H83.7082V167.391H95.6636V167.252H107.626V203.258ZM107.626 155.298H95.6636V119.563H107.626V155.298ZM107.626 107.602H95.6636V71.7349H107.543V83.7994H107.626V107.602ZM119.581 119.563H131.543V155.298H119.581V119.563ZM119.581 167.252H131.543V179.046H119.581V167.252ZM155.322 191.332H131.682V179.275H143.478V167.252H155.274L155.322 191.332ZM155.322 155.333H143.527V119.563H155.322V155.333ZM155.322 95.6003H143.527V107.637H119.581V83.7994H131.474V95.5933H143.429V83.5704H131.543V71.7349H155.322V95.6003ZM179.379 203.293H167.278V167.287H179.379V203.293ZM203.289 203.293H191.334V167.287H203.289V203.293ZM203.289 155.333H167.278V119.563H203.289V155.333ZM203.289 107.637H167.278V95.6488H203.289V107.637ZM203.289 83.73H167.278V71.7349H203.289V83.73Z" fill="black"/>
              <path d="M705.539 155.437H693.584V167.391H705.539V155.437Z" fill="black"/>
              <path d="M753.048 107.644V107.609H729.457V119.563H741.419V143.477H717.501V131.523H729.457V119.563H717.501V131.523H705.539V155.437H710.84H717.501H741.419V167.391H753.374V107.644H753.048Z" fill="black"/>
              <path d="M801.209 119.563V107.609H765.336V119.563H777.292V155.437H765.336V167.391H801.209V155.437H789.247V119.563H801.209Z" fill="black"/>
              <path d="M406.586 131.523V119.563H442.459V107.609H406.586H394.624V143.477H406.586H442.459V155.437H454.414V131.523H442.459H406.586Z" fill="black"/>
              <path d="M442.16 155.437H394.624V167.391H442.16V155.437Z" fill="black"/>
              <path d="M275.042 119.563H286.998V155.437H275.042V167.391H310.915V155.437H298.96V119.563H310.915V107.609H275.042V119.563Z" fill="black"/>
              <path d="M454.414 119.563H478.332V167.391H490.294V119.563H514.489V107.609H454.414V119.563Z" fill="black"/>
              <path d="M526.306 119.563H538.192V155.437H526.306V167.391H562.04V155.437H550.154V119.563H562.04V107.609H526.306V119.563Z" fill="black"/>
              <path d="M585.958 107.609H574.002V155.437V167.099V167.391H609.875V155.437H585.958V107.609Z" fill="black"/>
              <path d="M633.793 107.609H621.831V155.437V167.099V167.391H657.711V155.437H633.793V107.609Z" fill="black"/>
              <path d="M334.833 119.563V107.609H322.878V119.563V131.523V167.391H334.833V131.523H346.795V119.563H334.833Z" fill="black"/>
              <path d="M358.744 131.523H346.788V143.477H358.744V131.523Z" fill="black"/>
              <path d="M370.706 143.477H358.751V155.437H370.706V167.391H382.668V155.437V143.477V107.609H370.706V143.477Z" fill="black"/>
            </svg>
            <h1 class="title">Login Success</h1>
            <p class="text">
              Signed in via your OIDC provider. You can now close the window and
              start using Instill AI.
            </p>
          </div>
        </div>
      </div>
    </div>
  </body>
</html>
`
