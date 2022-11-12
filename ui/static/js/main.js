document.addEventListener('DOMContentLoaded', () => {
     /* Open Account Modal */
     const btnOpenAccountModal = document.querySelector('#open__account__modal');
     const openAccountModal = document.querySelector('#acc__modal');
     let openAccountModalChecker = 0;
     btnOpenAccountModal.addEventListener('click', () => {
          openAccountModalChecker++;
          if (openAccountModalChecker % 2 === 1) {
               openAccountModal.style.display = 'block';
          } else {
               openAccountModal.style.display = 'none';
          }
     });
     /* Change Data in My Profile */
     const postsList = document.querySelector('.posts');
     const changeDataList = document.querySelector('.change__data');
     const openChangeDataBtn = document.querySelector('#open__change');
     const closeChangeDataBtn = document.querySelector('#close__change');
     openChangeDataBtn.addEventListener('click', (e) => {
          e.preventDefault();
          postsList.classList.remove('show', 'fade');
          changeDataList.classList.add('show', 'fade');
     });
     closeChangeDataBtn.addEventListener('click', () => {
          changeDataList.classList.remove('show', 'fade');
          postsList.classList.add('show', 'fade');
     });


});